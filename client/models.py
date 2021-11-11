from pydantic import BaseModel
from typing import List


class AddressQuery(BaseModel):
    address: str


class AddressQueryList(BaseModel):
    __root__: List[AddressQuery]
