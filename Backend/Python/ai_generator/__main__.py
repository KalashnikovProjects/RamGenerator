import config
import image_generator


if __name__ == '__main__':
    api = image_generator.Text2ImageAPI('https://api-key.fusionbrain.ai/', config.Kandinsky.key, config.Kandinsky.secret_key)
    model_id = api.get_model()
    uuid = api.generate("Sun in sky", model_id)
    images = api.check_generation(uuid)
    print(images)