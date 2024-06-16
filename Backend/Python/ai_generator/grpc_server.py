import logging
from base64 import b64decode
from concurrent import futures

import grpc

import ai_generators
import config
import ram_generator_pb2
import ram_generator_pb2_grpc


def generate_error_handler(status_code, error: str):
    def error_handler(request, context):
        context.abort(status_code, error)

    return grpc.unary_unary_rpc_method_handler(error_handler)


class AuthInterceptor(grpc.ServerInterceptor):
    def intercept_service(self, continuation, handler_call_details):
        metadata = dict(handler_call_details.invocation_metadata)
        auth_token = metadata.get('authorization')
        if not auth_token or not auth_token.startswith('Bearer ') or auth_token[7:] != config.gRPC.SECRET_TOKEN:
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
                                                  system_instructions=config.PROMPTS.PERMANENT_START_PROMPT_GENERATOR_PROMPT)

        res = generator.generate(request.user_prompt)
        return ram_generator_pb2.ImagePrompt(prompt=res)

    @staticmethod
    def GenerateHybridPrompt(request, context, **kwargs):
        logging.info(f"Генерация гибрида промпта, промпт:{request.user_prompt}")

        generator = ai_generators.PromptGenerator(api_key=config.GEMINI.API_KEY,
                                                  model_name=config.GEMINI.MODEL,
                                                  safety_settings=config.GEMINI.safety_settings,
                                                  system_instructions=config.PROMPTS.PERMANENT_HYBRID_PROMPT_GENERATOR_PROMPT)
        images = [{"mime_type": "image/jpeg", "data": b64decode(im)} for im in request.ram_images]
        res = generator.generate(request.user_prompt, images)
        return ram_generator_pb2.ImagePrompt(prompt=res)

    @staticmethod
    def GenerateImage(request, context, **kwargs):
        logging.info(f"Генерация изображения, промпт:{request.prompt}")
        api = ai_generators.ImageGenerator(config.KANDINSKY.ENDPOINT, config.KANDINSKY.KEY,
                                           config.KANDINSKY.SECRET_KEY)
        model_id = api.get_model()
        uuid = api.generate(f"{config.PROMPTS.PERMANENT_IMAGE_PROMPT}, {request.prompt}", model_id)

        try:
            image = api.check_generation(uuid)
            return ram_generator_pb2.GenerateImageResponse(ram_image=image)
        except ai_generators.ImageGenerationTimeoutError:
            context.abort(grpc.StatusCode.DEADLINE_EXCEEDED, f"The waiting time for image generation has been exceeded")
        except Exception as e:
            context.abort(grpc.StatusCode.INTERNAL, f"Internal server error: {str(e)}")

        return context

    @classmethod
    def serve(cls):
        port = config.gRPC.PORT

        server = grpc.server(futures.ThreadPoolExecutor(max_workers=30),
                             interceptors=[AuthInterceptor()])
        ram_generator_pb2_grpc.add_RamGeneratorServicer_to_server(cls(), server)
        server.add_insecure_port(f'[::]:{port}')
        server.start()
        logging.info(f"RamGenerator gRPC started at port {port}")
        server.wait_for_termination()
