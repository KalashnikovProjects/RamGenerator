import os

from dotenv import load_dotenv
load_dotenv("../../../.env")


class Kandinsky:
    key = os.getenv("kandinsky_key")
    secret_key = os.getenv("kandinsky_secret_key")
