import logging

import google.generativeai as genai
import json
import requests
import time
from retry import retry
from typing_extensions import TypedDict

from . import rate_limiters
from . import config


class GeminiCensorshipError(Exception):
    pass


class GeminiBugError(Exception):
    pass


class ImageCensorshipError(Exception):
    pass


class ImageGenerationUnavailableError(Exception):
    pass


class ImageGenerationInternalError(Exception):
    pass


class NoRamError(Exception):
    pass


DescriptionResponse = TypedDict('DescriptionResponse', {'contains_ram': bool, 'description': str}, total=True)

PromptResponse = TypedDict('PromptResponse', {'contains_swear': bool, 'prompt': str}, total=True)


class PromptGenerator:
    def __init__(self,
                 system_instructions: str,
                 max_output_tokens: int = None,
                 model_name: str = config.GEMINI.MODEL,
                 safety_settings=None):

        self.model = genai.GenerativeModel(
            model_name=model_name,
            safety_settings=safety_settings,
            system_instruction=system_instructions,
            generation_config=genai.GenerationConfig(candidate_count=1, max_output_tokens=max_output_tokens)
        )

    @retry(tries=3, delay=2)
    @rate_limiters.api_rate_limiter_with_que(rate_limit=config.GEMINI.RATE_LIMIT)
    def generate(self, text: str, images: list[dict[str, bytes | str]] = None, generation_config=None) -> str:
        inp = [text, *images] if images else text

        res = self.model.generate_content(inp, generation_config=generation_config)

        if not res.parts:
            if res.candidates[0].finish_reason == 3:
                raise GeminiCensorshipError
            else:
                raise GeminiBugError
        return res.text.strip()

    @staticmethod
    def configure():
        genai.configure(transport=config.GEMINI.TRANSPORT, api_key=config.GEMINI.API_KEY,
                        client_options={"api_endpoint": config.GEMINI.ENDPOINT})


class ImageGenerationTimeoutError(TimeoutError):
    pass


class ImageGenerator:
    def __init__(self, url, api_key):
        self.URL = url
        self.API_KEY = api_key

    # @retry(tries=3, delay=2)
    def generate(self, prompt, width=512, height=512):
        data = {
            "prompt": prompt,
            "width": width,
            "height": height,
        }
        try:
            response = requests.post(
                self.URL,
                json=data,
                headers={"Authorization": f"Bearer {self.API_KEY}",
                         "Content-Type": "application/json"},
                timeout=100,
            )
            if response.status_code == 200:
                return response.json()["image"]
            else:
                raise ImageGenerationInternalError
        except TimeoutError as e:
            raise ImageGenerationTimeoutError