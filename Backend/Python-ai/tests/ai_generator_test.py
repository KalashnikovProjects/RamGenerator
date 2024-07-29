import unittest
import logging
import grpc_testing

from ..ai_generator import grpc_server
from ..ai_generator.proto_generated import ram_generator_pb2


# TODO:
class SampleServiceTest(unittest.TestCase):
    def setUp(self):
        logging.info( f"=== Method: { self._testMethodName } =======" )
        servicers = {
            ram_generator_pb2.DESCRIPTOR.services_by_name['SampleService']: grpc_server.RamGeneratorServicer()
        }

        self.test_server = grpc_testing.server_from_dictionary(
            servicers, grpc_testing.strict_real_time())

    def new_prompt_generator_test(self):
        # TODO
        self.assertEqual(True, False)

    def hybrid_prompt_generator_test(self):
        # TODO
        self.assertEqual(True, False)

    def image_generator_test(self):
        # TODO
        ...


if __name__ == '__main__':
    unittest.main()
