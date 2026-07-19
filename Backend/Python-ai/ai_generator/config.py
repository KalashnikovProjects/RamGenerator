import logging
import os
import yaml
from dotenv import load_dotenv
from google.generativeai.types import HarmCategory, HarmBlockThreshold


ROOT_PATH = os.getenv("ROOT_PATH", "../..")

load_dotenv(f"{ROOT_PATH}/.env")  # For working without docker
with open(f"{ROOT_PATH}/config.yaml") as f:
    yaml_config = yaml.load(f, Loader=yaml.FullLoader)

LOG_LEVEL = logging.DEBUG


class GRPC:
    PORT = os.getenv("GRPC_PORT")
    SECRET_TOKEN = os.getenv("GRPC_SECRET_TOKEN")


class PROMPTS:
    BASE_START_PROMPT = yaml_config["prompts"]["base_start_prompt"]
    BASE_HYBRID_PROMPT = yaml_config["prompts"]["base_hybrid_prompt"]

    BASE_IMAGE_PROMPT = yaml_config["prompts"]["base_image_prompt"]

    BASE_DESCRIPTION_PROMPT = yaml_config["prompts"]["base_description_prompt"]


class IMAGE_GENERATION:
    API_KEY = os.getenv("IMAGE_GENERATOR_API_KEY")
    ENDPOINT = yaml_config["image_generation"]["endpoint"]
    IMAGE_SIDE = yaml_config["image_generation"]["image_side"]


class GEMINI:
    API_KEY = os.getenv("GEMINI_API_KEY")
    MODEL = os.getenv("GEMINI_MODEL") or yaml_config["gemini"]["model"]
    RATE_LIMIT = yaml_config["gemini"]["rate_limit"]

    proxy = yaml_config["gemini"]["proxy"]
    if proxy == "true":
        ENDPOINT = yaml_config["gemini"]["proxy_endpoint"]
        TRANSPORT = yaml_config["gemini"]["proxy_transport"]
    else:
        ENDPOINT = None
        TRANSPORT = None

    SAFETY_SETTINGS = {
        HarmCategory.HARM_CATEGORY_HARASSMENT: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE,
        HarmCategory.HARM_CATEGORY_HATE_SPEECH: HarmBlockThreshold.BLOCK_ONLY_HIGH,
        HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT: HarmBlockThreshold.BLOCK_LOW_AND_ABOVE,
        HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE
    }
