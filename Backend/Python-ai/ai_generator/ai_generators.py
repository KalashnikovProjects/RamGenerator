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


class NoRamError(Exception):
    pass


# Dict keys on russian for russia model response
DescriptionResponse = TypedDict('DescriptionResponse', {'есть баран': bool, 'краткое описание': str}, total=True)


# Dict keys on russian for russia model response
PromptResponse = TypedDict('PromptResponse', {'есть мат': bool, 'запрос': str}, total=True)


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
    def __init__(self, url, api_key, secret_key):
        self.URL = url
        self.AUTH_HEADERS = {
            'X-Key': f'Key {api_key}',
            'X-Secret': f'Secret {secret_key}',
        }

    def get_model(self):
        response = requests.get(self.URL + 'key/api/v1/models', headers=self.AUTH_HEADERS)
        data = response.json()
        return data[0]['id']

    def generate(self, prompt, style, model, width, height, images=1):
        params = {
            "type": "GENERATE",
            "style": style,
            "numImages": images,
            "negativePromptUnclip": config.PROMPTS.IMAGE_NEGATIVE_PROMPT,
            "width": width,
            "height": height,
            "generateParams": {
                "query": f"{prompt}"
            }
        }

        data = {
            'model_id': (None, model),
            'params': (None, json.dumps(params), 'application/json')
        }
        response = requests.post(self.URL + 'key/api/v1/text2image/run', headers=self.AUTH_HEADERS, files=data)
        data = response.json()
        return data['uuid']

    def check_generation(self, request_id, attempts=20, delay=10):
        for attempt in range(attempts):
            response = requests.get(self.URL + 'key/api/v1/text2image/status/' + request_id, headers=self.AUTH_HEADERS)
            data = response.json()
            if data['status'] == 'DONE':
                return data['images'][0]
            if data["censored"] == "FAIL":
                if data["censored"]:
                    raise ImageCensorshipError
                else:
                    raise ImageGenerationUnavailableError
            time.sleep(delay)
        raise ImageGenerationTimeoutError(f'Image generation failed after {attempts} attempts')
