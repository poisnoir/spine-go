import msgpack

class Mad:
    def __init__(self, type_hint=None):
        self.type_hint = type_hint or "any"

    def code(self):
        # In spine-go, Code() returns a string/byte slice identifying the type.
        # We'll use the type name as a simple identifier.
        return self.type_hint.encode('utf-8')

    def encode(self, data):
        return msgpack.packb(data)

    def decode(self, data):
        return msgpack.unpackb(data)

    def get_required_size(self, data):
        # msgpack.packb returns a bytes object, so we can just return its length.
        # However, for the sake of following Go logic where this is used for buffer sizing:
        return len(self.encode(data))

def new_mad(type_hint=None):
    return Mad(type_hint)
