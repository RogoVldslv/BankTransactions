import asyncio
import json
import os
from typing import Union

from aiogram import Bot
from redis.asyncio import from_url

async def make_text(amount: int, message: Union[None,str]) -> str:
    if message:
        res = f"Тебе пришел перевод в размере {amount}\nСообщение отправителя: {message}"
    else:
        res = f"Тебе пришел перевод в размере {amount}"
    return res
async def main():
    bot = Bot(token=os.getenv("BOT_TOKEN"))
    r = from_url(os.getenv("REDIS_URL"), decode_responses=True)

    print("Бот-воркер запущен. Ожидание уведомлений...")

    for _ in range(5):
        try:
            result = await r.brpop(os.getenv("REDIS_QUEUE_NAME"), timeout=0)
            if result:
                _, data = result
                payload = json.loads(data)

                chat_id = payload.get("tg_id")
                text = payload.get("message")
                amount = payload.get("amount")

                if chat_id and text:
                    await bot.send_message(chat_id, await make_text(amount, text))
                    print(f"Сообщение отправлено пользователю {chat_id}")
        except Exception as e:
            print(f"Ошибка запуска : {e}")
            await asyncio.sleep(10)
    await bot.session.close()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("Бот остановлен")