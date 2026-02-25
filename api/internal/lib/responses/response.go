package response

import "github.com/gin-gonic/gin"

func Error(error string) gin.H {
	return gin.H{"status": "error", "msg": error}
}

func Ok() gin.H {
	return gin.H{"status": "ok"}
}
