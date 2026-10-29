package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"BankTransactions/internal/http_server/handlers"
	"BankTransactions/internal/models"
	"BankTransactions/internal/storage/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	sqliteDriver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type fixture struct {
	router        *gin.Engine
	db            *gorm.DB
	notifications chan []string
}

// A small RESP server records LPUSH; no external Redis is needed.
func redisStub(t *testing.T) (*redis.Client, chan []string) {
	t.Helper()
	notifications := make(chan []string, 10)
	client := redis.NewClient(&redis.Options{
		Protocol:        2,
		DisableIdentity: true,
		Dialer: func(_ context.Context, _, _ string) (net.Conn, error) {
			clientConn, serverConn := net.Pipe()
			go serveRedisStub(serverConn, notifications)
			return clientConn, nil
		},
	})
	t.Cleanup(func() { client.Close() })
	return client, notifications
}

func serveRedisStub(conn net.Conn, notifications chan<- []string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "*")))
		if err != nil {
			return
		}
		args := make([]string, n)
		for i := range args {
			line, err = reader.ReadString('\n')
			if err != nil {
				return
			}
			size, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "$")))
			if err != nil || size < 0 {
				return
			}
			buf := make([]byte, size+2)
			if _, err = io.ReadFull(reader, buf); err != nil {
				return
			}
			args[i] = string(buf[:size])
		}
		if len(args) == 0 {
			return
		}
		switch strings.ToUpper(args[0]) {
		case "HELLO":
			_, err = io.WriteString(conn, "-ERR unknown command\r\n")
		case "LPUSH":
			notifications <- args
			_, err = io.WriteString(conn, ":1\r\n")
		default:
			_, err = io.WriteString(conn, "+OK\r\n")
		}
		if err != nil {
			return
		}
	}
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("REDIS_QUEUE_NAME", "test_transfers")
	db, err := gorm.Open(sqliteDriver.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		t.Fatal(err)
	}
	client, notifications := redisStub(t)
	stg := &sqlite.Storage{DB: db}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := gin.New()
	router.POST("/account", handlers.CreatePostAccount(log, stg))
	router.POST("/transaction", handlers.TransferMoney(log, stg, client))
	return fixture{router, db, notifications}
}

func (f fixture) post(path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	f.router.ServeHTTP(w, req)
	return w
}

func expectResponse(t *testing.T, w *httptest.ResponseRecorder, status int, result, message string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("HTTP status = %d, want %d; body: %s", w.Code, status, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body["status"] != result {
		t.Errorf("status = %v, want %q", body["status"], result)
	}
	if message != "" && !strings.Contains(fmt.Sprint(body["msg"]), message) {
		t.Errorf("message = %v, want %q", body["msg"], message)
	}
}

func createAccount(t *testing.T, f fixture, name string, balance, passport, tgID int) models.Account {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"balance":%d,"passport_id":%d,"tg_id":%d}`, name, balance, passport, tgID)
	expectResponse(t, f.post("/account", body), http.StatusOK, "ok", "")
	var a models.Account
	if err := f.db.Where("passport_id = ?", passport).First(&a).Error; err != nil {
		t.Fatal(err)
	}
	return a
}

func assertState(t *testing.T, f fixture, a, b models.Account, balanceA, balanceB int64, transactions int64) {
	t.Helper()
	for _, item := range []struct {
		id   uint
		want int64
	}{{a.ID, balanceA}, {b.ID, balanceB}} {
		var got models.Account
		if err := f.db.First(&got, item.id).Error; err != nil {
			t.Fatal(err)
		}
		if got.Balance != item.want {
			t.Errorf("account %d balance = %d, want %d", item.id, got.Balance, item.want)
		}
	}
	var count int64
	if err := f.db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != transactions {
		t.Errorf("transactions = %d, want %d", count, transactions)
	}
}

func TestAccount(t *testing.T) {
	f := newFixture(t)
	a := createAccount(t, f, "Алиса", 100, 12345678, 101)
	if a.Name != "Алиса" || a.Balance != 100 || a.TelegramID != 101 {
		t.Fatalf("stored account: %+v", a)
	}
	expectResponse(t, f.post("/account", `{"name":"Bob","balance":30,"passport_id":12345678,"tg_id":102}`), 400, "error", "same passport")
	for _, tc := range []struct{ name, body string }{
		{"malformed JSON", `{`},
		{"missing name", `{"passport_id":12345679,"tg_id":102}`},
		{"missing telegram", `{"name":"Bob","passport_id":12345679}`},
		{"short passport", `{"name":"Bob","passport_id":1234567,"tg_id":102}`},
		{"long passport", `{"name":"Bob","passport_id":100000000,"tg_id":102}`},
		{"wrong balance type", `{"name":"Bob","balance":"100","passport_id":12345679,"tg_id":102}`},
	} {
		t.Run(tc.name, func(t *testing.T) { expectResponse(t, f.post("/account", tc.body), 400, "error", "") })
	}
	var count int64
	if err := f.db.Model(&models.Account{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("accounts = %d, want 1", count)
	}
}

func TestTransfer(t *testing.T) {
	f := newFixture(t)
	a := createAccount(t, f, "Alice", 100, 12345678, 101)
	b := createAccount(t, f, "Bob", 20, 87654321, 202)
	body := fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":40,"message":"hello"}`, a.ID, b.ID)
	expectResponse(t, f.post("/transaction", body), 200, "ok", "")
	assertState(t, f, a, b, 60, 60, 1)
	var tx models.Transaction
	if err := f.db.First(&tx).Error; err != nil {
		t.Fatal(err)
	}
	if tx.FromUserId != int(a.ID) || tx.ToUserId != int(b.ID) || tx.Amount != 40 || tx.Message != "hello" {
		t.Errorf("stored transaction: %+v", tx)
	}
	select {
	case args := <-f.notifications:
		if len(args) != 3 || args[1] != "test_transfers" {
			t.Fatalf("LPUSH: %v", args)
		}
		var payload struct {
			From    int    `json:"from_user_id"`
			To      int    `json:"to_user_id"`
			Amount  int    `json:"amount"`
			Message string `json:"message"`
			TgID    int    `json:"tg_id"`
		}
		if err := json.Unmarshal([]byte(args[2]), &payload); err != nil {
			t.Fatal(err)
		}
		if payload.From != int(a.ID) || payload.To != int(b.ID) || payload.Amount != 40 || payload.Message != "hello" || payload.TgID != 202 {
			t.Errorf("notification: %+v", payload)
		}
	case <-time.After(time.Second):
		t.Fatal("notification not received")
	}
}

func TestTransferFailuresLeaveStateUnchanged(t *testing.T) {
	f := newFixture(t)
	a := createAccount(t, f, "Alice", 100, 12345678, 101)
	b := createAccount(t, f, "Bob", 20, 87654321, 202)
	cases := []struct {
		name, body string
		status     int
		message    string
	}{
		{"malformed JSON", `{`, 400, "bad data"},
		{"missing amount", fmt.Sprintf(`{"from_id":%d,"to_id":%d}`, a.ID, b.ID), 400, ""},
		{"zero amount", fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":0}`, a.ID, b.ID), 400, ""},
		{"negative amount", fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":-1}`, a.ID, b.ID), 400, ""},
		{"message too long", fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":1,"message":%q}`, a.ID, b.ID, strings.Repeat("x", 33)), 400, ""},
		{"missing sender", fmt.Sprintf(`{"from_id":999,"to_id":%d,"amount":1}`, b.ID), 404, "account from"},
		{"missing recipient", fmt.Sprintf(`{"from_id":%d,"to_id":999,"amount":1}`, a.ID), 404, "account to"},
		{"insufficient funds", fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":101}`, a.ID, b.ID), 422, "insufficient funds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectResponse(t, f.post("/transaction", tc.body), tc.status, "error", tc.message)
			assertState(t, f, a, b, 100, 20, 0)
			select {
			case msg := <-f.notifications:
				t.Fatalf("unexpected notification: %v", msg)
			default:
			}
		})
	}
}

func TestTransferRollsBackWhenSavingFails(t *testing.T) {
	f := newFixture(t)
	a := createAccount(t, f, "Alice", 100, 12345678, 101)
	b := createAccount(t, f, "Bob", 20, 87654321, 202)
	if err := f.db.Exec(`CREATE TRIGGER reject_transfer BEFORE INSERT ON transactions BEGIN SELECT RAISE(ABORT, 'test failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":40}`, a.ID, b.ID)
	expectResponse(t, f.post("/transaction", body), 500, "error", "internal error")
	assertState(t, f, a, b, 100, 20, 0)
	select {
	case msg := <-f.notifications:
		t.Fatalf("unexpected notification: %v", msg)
	default:
	}
}

func TestConcurrentTransfersCannotOverdraw(t *testing.T) {
	f := newFixture(t)
	a := createAccount(t, f, "Alice", 100, 12345678, 101)
	b := createAccount(t, f, "Bob", 20, 87654321, 202)
	body := fmt.Sprintf(`{"from_id":%d,"to_id":%d,"amount":70}`, a.ID, b.ID)

	start := make(chan struct{})
	responses := make(chan *httptest.ResponseRecorder, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			responses <- f.post("/transaction", body)
		}()
	}
	close(start)

	success, rejected := 0, 0
	for i := 0; i < 2; i++ {
		w := <-responses
		switch w.Code {
		case http.StatusOK:
			expectResponse(t, w, http.StatusOK, "ok", "")
			success++
		case http.StatusUnprocessableEntity:
			expectResponse(t, w, http.StatusUnprocessableEntity, "error", "insufficient funds")
			rejected++
		case http.StatusInternalServerError:
			// SQLite may reject a concurrent write with SQLITE_BUSY.
			expectResponse(t, w, http.StatusInternalServerError, "error", "internal error")
			rejected++
			t.Log("concurrent transfer was rejected with HTTP 500")
		default:
			t.Errorf("unexpected response: HTTP %d: %s", w.Code, w.Body.String())
		}
	}
	if success != 1 || rejected != 1 {
		t.Fatalf("successful transfers = %d, rejected transfers = %d; want 1 each", success, rejected)
	}
	assertState(t, f, a, b, 30, 90, 1)
	select {
	case <-f.notifications:
	default:
		t.Fatal("successful transfer did not send a notification")
	}
	select {
	case msg := <-f.notifications:
		t.Fatalf("unexpected extra notification: %v", msg)
	default:
	}
}
