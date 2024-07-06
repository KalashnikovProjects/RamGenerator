import logging
from concurrent import futures

import grpc

from . import ai_generators
from . import config
from .proto_generated import ram_generator_pb2
from .proto_generated import ram_generator_pb2_grpc


def generate_error_handler(status_code, error: str):
    def error_handler(request, context):
        context.abort(status_code, error)

    return grpc.unary_unary_rpc_method_handler(error_handler)


class AuthInterceptor(grpc.ServerInterceptor):
    def intercept_service(self, continuation, handler_call_details):
        metadata = dict(handler_call_details.invocation_metadata)
        auth_token = metadata.get('authorization')
        if not auth_token or not auth_token.startswith('Bearer ') or auth_token[7:] != config.GRPC.SECRET_TOKEN:
            logging.debug("Failed authorization")
            return generate_error_handler(grpc.StatusCode.UNAUTHENTICATED, "Access denied. Authentication required.")
        else:
            logging.debug("Success authorization")
            return continuation(handler_call_details)


class RamGeneratorServer(ram_generator_pb2_grpc.RamGenerator):
    @staticmethod
    def GenerateStartPrompt(request, context, **kwargs):
        logging.info(f"Генерация промпта, промпт:{request.user_prompt}")

        generator = ai_generators.PromptGenerator(api_key=config.GEMINI.API_KEY,
                                                  model_name=config.GEMINI.MODEL,
                                                  safety_settings=config.GEMINI.safety_settings,
                                                  system_instructions=config.PROMPTS.BASE_START_PROMPT_GENERATOR_PROMPT)
        prompt = f"Напиши промпт для генерации изображения нового барана пользователя. \nЗапрос пользователя: {request.user_prompt}"
        try:
            res = generator.generate(prompt)
            return ram_generator_pb2.RamImagePrompt(prompt=res)
        except ai_generators.GeminiCensorshipError:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "User prompt contains illegal content")
        except Exception as e:
            context.abort(grpc.StatusCode.INTERNAL, f"Internal server error: {str(e)}")

    @staticmethod
    def GenerateHybridPrompt(request, context, **kwargs):
        logging.info(f"Генерация гибрида промпта, промпт:{request.user_prompt}")

        generator = ai_generators.PromptGenerator(api_key=config.GEMINI.API_KEY,
                                                  model_name=config.GEMINI.MODEL,
                                                  safety_settings=config.GEMINI.safety_settings,
                                                  system_instructions=config.PROMPTS.BASE_HYBRID_PROMPT_GENERATOR_PROMPT)
        rams = '\n'.join(request.ram_descriptions)
        prompt = f"Напиши промпт для генерации изображения нового барана пользователя. \nЗапрос пользователя: {request.user_prompt}\nОписание баранов пользователя: \n{rams}"
        try:
            res = generator.generate(prompt, [])
            return ram_generator_pb2.RamImagePrompt(prompt=res)
        except ai_generators.GeminiCensorshipError:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "User prompt or descriptions contains illegal content")
        except Exception as e:
            context.abort(grpc.StatusCode.INTERNAL, f"Internal server error: {str(e)}")

    @staticmethod
    def GenerateImage(request, context, **kwargs):
        logging.info(f"Генерация изображения, промпт:{request.prompt}")
        api = ai_generators.ImageGenerator(config.KANDINSKY.ENDPOINT, config.KANDINSKY.KEY,
                                           config.KANDINSKY.SECRET_KEY)
        model_id = api.get_model()
        uuid = api.generate(f"{config.PROMPTS.BASE_IMAGE_PROMPT}, {request.prompt}", request.style, model_id)

        try:
            image = api.check_generation(uuid)
            return ram_generator_pb2.RamImage(image=image)
        except ai_generators.ImageGenerationTimeoutError:
            context.abort(grpc.StatusCode.DEADLINE_EXCEEDED, f"The waiting time for image generation has been exceeded")
        except Exception as e:
            context.abort(grpc.StatusCode.INTERNAL, f"Internal server error: {str(e)}")

        return context

    @staticmethod
    def GenerateDescription(request, context, **kwargs):
        logging.info(f"Генерация описания")

        generator = ai_generators.PromptGenerator(api_key=config.GEMINI.API_KEY,
                                                  model_name=config.GEMINI.MODEL,
                                                  safety_settings=config.GEMINI.safety_settings,
                                                  system_instructions=config.PROMPTS.BASE_DESCRIPTION_PROMPT)
        try:
            res = generator.generate("Напиши описание для изображения барана", [{"file_uri": request.url}])
            return ram_generator_pb2.RamDescription(description=res)
        except ai_generators.GeminiCensorshipError:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "The image contains illegal content")
        except Exception as e:
            context.abort(grpc.StatusCode.INTERNAL, f"Internal server error: {str(e)}")

    @classmethod
    def serve(cls):
        port = config.GRPC.PORT

        server = grpc.server(futures.ThreadPoolExecutor(max_workers=30),
                             interceptors=[AuthInterceptor()])
        ram_generator_pb2_grpc.add_RamGeneratorServicer_to_server(cls(), server)
        server.add_insecure_port(f'[::]:{port}')
        server.start()
        logging.info(f"RamGenerator gRPC started at port {port}")
        server.wait_for_termination()
