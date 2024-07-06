import logging
import os
from dotenv import load_dotenv
from google.generativeai.types import HarmCategory, HarmBlockThreshold

load_dotenv("../../.env")  # Такой путь чтобы работало без докера

LOG_LEVEL = logging.DEBUG


class GRPC:
    PORT = os.getenv("GRPC_PORT")
    SECRET_TOKEN = os.getenv("GRPC_SECRET_TOKEN")


class PROMPTS:
    BASE_START_PROMPT_GENERATOR_PROMPT = ("Ты пишешь промты для генерации баранов с помощью нейросети Kandinsky на основе запроса"
                                          "пользователя. Ты не имеешь право задавать в промпте параметры чего-либо кроме внешнего "
                                          "вида барана. Помимо требований пользователя ты должен добавить "
                                          "от себя 1 случайный признак для барана. (мутацию)")
    BASE_HYBRID_PROMPT_GENERATOR_PROMPT = ("Ты пишешь промт для нейросети Kandinsky на основе запроса пользователя и описания баранов в стаде. "
                                           "Нейросеть для генерации изображений рисует различных баранов, "
                                           "ты не имеешь право задавать в промпте параметры чего-либо кроме внешнего "
                                           "вида барана. Ты пишешь промпт на основе запроса пользователя и в большей "
                                           "степени на основе описания других его баранов, ты должен добавить ровно"
                                           "1 случайный признак от себя (мутацию), всё остальное на основе запроса или других баранов.")

    BASE_IMAGE_PROMPT = "баран, лужайка на фоне, ясная погода, баран на переднем плане, "
    IMAGE_NEGATIVE_PROMPT = "яркие цвета, кислотность, высокая контрастность"

    BASE_DESCRIPTION_PROMPT = "Напиши максимально кратное описание барана на этом изображении (не более 6 слов)."


class KANDINSKY:
    ENDPOINT = 'https://api-key.fusionbrain.ai/'
    KEY = os.getenv("KANDINSKY_KEY")
    SECRET_KEY = os.getenv("KANDINSKY_SECRET_KEY")


class GEMINI:
    API_KEY = os.getenv("GEMINI_API_KEY")
    MODEL = "models/gemini-1.5-flash-latest"
    RATE_LIMIT = 15
    DEFAULT_RESPONSE_LEN = 50

    proxy = True
    if proxy:
        ENDPOINT = "cheery-baklava-3e2f26.netlify.app"
        TRANSPORT = "rest"
    else:
        ENDPOINT = None
        TRANSPORT = None

    safety_settings = {
        HarmCategory.HARM_CATEGORY_HARASSMENT: HarmBlockThreshold.BLOCK_ONLY_HIGH,
        HarmCategory.HARM_CATEGORY_HATE_SPEECH: HarmBlockThreshold.BLOCK_NONE,
        HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT: HarmBlockThreshold.BLOCK_ONLY_HIGH,
        HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT: HarmBlockThreshold.BLOCK_ONLY_HIGH
    }
