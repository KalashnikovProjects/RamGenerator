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
    __slots__ = ("user_prompt", "ram_descriptions")
    USER_PROMPT_FIELD_NUMBER: _ClassVar[int]
    RAM_DESCRIPTIONS_FIELD_NUMBER: _ClassVar[int]
    user_prompt: str
    ram_descriptions: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, user_prompt: _Optional[str] = ..., ram_descriptions: _Optional[_Iterable[str]] = ...) -> None: ...

class RamImagePrompt(_message.Message):
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

class RamImage(_message.Message):
    __slots__ = ("image",)
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    image: str
    def __init__(self, image: _Optional[str] = ...) -> None: ...

class RamImageUrl(_message.Message):
    __slots__ = ("url",)
    URL_FIELD_NUMBER: _ClassVar[int]
    url: str
    def __init__(self, url: _Optional[str] = ...) -> None: ...

class RamDescription(_message.Message):
    __slots__ = ("description",)
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    description: str
    def __init__(self, description: _Optional[str] = ...) -> None: ...
