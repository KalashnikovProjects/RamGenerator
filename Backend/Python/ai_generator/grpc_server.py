from ..proto_generated import ram_generator_pb2, ram_generator_pb2_grpc
import image_generator
import config


class RamGeneratorServer(ram_generator_pb2_grpc.RamGenerator):
    async def GenerateImage(self, request, context):
        api = image_generator.Text2ImageAPI('https://api-key.fusionbrain.ai/', config.Kandinsky.key,
                                            config.Kandinsky.secret_key)
        model_id = api.get_model()
        uuid = api.generate(request.prompt, model_id)
        images = api.check_generation(uuid)
        return ram_generator_pb2.GenerateImageResponse(ram_image=images[0])

    async def GenerateStartPrompt(self, request, context):
        return ram_generator_pb2.ImagePrompt(...)

    async def GenerateHybridPrompt(self, request, context):
        return ram_generator_pb2.ImagePrompt(...)