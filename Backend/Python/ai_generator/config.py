import logging
import os
from dotenv import load_dotenv
from google.generativeai.types import HarmCategory, HarmBlockThreshold

load_dotenv("../../../.env")  # Такой путь чтобы работало без докера

LOG_LEVEL = logging.DEBUG


class gRPC:
    PORT = 50051
    SECRET_TOKEN = os.getenv("GRPC_SECRET_TOKEN")


class PROMPTS:
    PERMANENT_START_PROMPT_GENERATOR_PROMPT = ("Ты пишешь промт для нейросети Kandinsky на основе запроса "
                                               "пользователя. Нейросеть для генерации изображений рисует различных баранов, "
                                               "ты не имеешь право задавать в промпте параметры чего-либо кроме внешнего "
                                               "вида барана (фона, стиля). Помимо требований пользователя ты должен добавить "
                                               "от себя 1 случайный признак для барана.")
    PERMANENT_HYBRID_PROMPT_GENERATOR_PROMPT = ("Ты пишешь промт для нейросети Kandinsky на основе изображений. "
                                                "Нейросеть для генерации изображений рисует различных баранов, "
                                                "ты не имеешь право задавать в промпте параметры чего-либо кроме внешнего "
                                                "вида барана (фона, стиля). Ты делаешь промпт для гибрида нескольких "
                                                "изображений баранов. Ты должен объединить основные признаки "
                                                "баранов с изображений и добавить 1 случайный признак (мутацию).")

    PERMANENT_IMAGE_PROMPT = "баран, лужайка на фоне, ясно, баран на переднем плане, "  # TODO


class KANDINSKY:
    ENDPOINT = 'https://api-key.fusionbrain.ai/'
    KEY = os.getenv("KANDINSKY_KEY")
    SECRET_KEY = os.getenv("KANDINSKY_SECRET_KEY")


class GEMINI:
    API_KEY = os.getenv("GEMINI_API_KEY")
    MODEL = "models/gemini-1.5-flash-latest"
    RATE_LIMIT = 15

    proxy = True
    if proxy:
        ENDPOINT = "cheery-baklava-3e2f26.netlify.app"
        TRANSPORT = "rest"
    else:
        ENDPOINT = None
        TRANSPORT = None

    safety_settings = {
        HarmCategory.HARM_CATEGORY_HARASSMENT: HarmBlockThreshold.BLOCK_NONE,
        HarmCategory.HARM_CATEGORY_HATE_SPEECH: HarmBlockThreshold.BLOCK_NONE,
        HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT: HarmBlockThreshold.BLOCK_ONLY_HIGH,
        HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT: HarmBlockThreshold.BLOCK_NONE
    }
