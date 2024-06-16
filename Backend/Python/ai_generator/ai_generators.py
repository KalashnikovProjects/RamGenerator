import google.generativeai as genai

from retry import retry

import rate_limiters
import config

import json
import time
import requests


class PromptGenerator:
    def __init__(self, api_key: str,
                 system_instructions: str,
                 model_name: str = config.GEMINI.MODEL,
                 safety_settings=None):

        genai.configure(transport=config.GEMINI.TRANSPORT, api_key=api_key,
                        client_options={"api_endpoint": config.GEMINI.ENDPOINT})
        self.model = genai.GenerativeModel(
            model_name=model_name,
            safety_settings=safety_settings,
            system_instruction=system_instructions,
            generation_config=genai.GenerationConfig(candidate_count=1)
        )

    @retry(tries=3, delay=2)
    @rate_limiters.api_rate_limiter_with_que(rate_limit=config.GEMINI.RATE_LIMIT)
    def generate(self, text: str, images: list[bytes] = None) -> str:
        inp = [text, *images] if images else text
        res = self.model.generate_content(inp)

        return res.text.strip()


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

    def generate(self, prompt, model, images=1, width=1024, height=1024):
        # TODO: добавить style
        params = {
            "type": "GENERATE",
            "numImages": images,
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

    def check_generation(self, request_id, attempts=10, delay=10):
        for attempt in range(attempts):
            response = requests.get(self.URL + 'key/api/v1/text2image/status/' + request_id, headers=self.AUTH_HEADERS)
            data = response.json()
            if data['status'] == 'DONE':
                return data['images'][0]

            time.sleep(delay)
        raise ImageGenerationTimeoutError(f'Image generation failed after {attempts} attempts')
