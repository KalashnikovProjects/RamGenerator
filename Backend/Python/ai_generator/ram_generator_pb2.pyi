from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class GenerateStartPromptRequest(_message.Message):
    __slots__ = ("user_prompt",)
    USER_PROMPT_FIELD_NUMBER: _ClassVar[int]
    user_prompt: str
    def __init__(self, user_prompt: _Optional[str] = ...) -> None: ...

class GenerateHybridPromptRequest(_message.Message):
    __slots__ = ("user_prompt", "ram_images")
    USER_PROMPT_FIELD_NUMBER: _ClassVar[int]
    RAM_IMAGES_FIELD_NUMBER: _ClassVar[int]
    user_prompt: str
    ram_images: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, user_prompt: _Optional[str] = ..., ram_images: _Optional[_Iterable[str]] = ...) -> None: ...

class ImagePrompt(_message.Message):
    __slots__ = ("prompt",)
    PROMPT_FIELD_NUMBER: _ClassVar[int]
    prompt: str
    def __init__(self, prompt: _Optional[str] = ...) -> None: ...

class GenerateImageRequest(_message.Message):
    __slots__ = ("prompt", "style")
    PROMPT_FIELD_NUMBER: _ClassVar[int]
    STYLE_FIELD_NUMBER: _ClassVar[int]
    prompt: str
    style: str
    def __init__(self, prompt: _Optional[str] = ..., style: _Optional[str] = ...) -> None: ...

class GenerateImageResponse(_message.Message):
    __slots__ = ("ram_image",)
    RAM_IMAGE_FIELD_NUMBER: _ClassVar[int]
    ram_image: str
    def __init__(self, ram_image: _Optional[str] = ...) -> None: ...
