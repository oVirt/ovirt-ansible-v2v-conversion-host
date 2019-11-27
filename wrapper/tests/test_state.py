import json
import tempfile
import unittest
from wrapper.state import STATE, Disk


class TestState(unittest.TestCase):
    """ Tests state object, make sure it behaves like a proper singleton """

    def setUp(self):
        # Destroy any previous state
        STATE.reset()

    def test_dict(self):
        """ Make sure the access to internal dictionary works """
        self.assertEqual(STATE['disks'], [])
        self.assertEqual(STATE['internal']['disk_ids'], {})
        # check -- change -- check
        self.assertEqual(STATE.failed, False)
        STATE.failed = True
        self.assertEqual(STATE.failed, True)

    def test_state(self):
        state1 = STATE
        state2 = STATE
        # Internal dictionary
        key = 'abcdef'
        value = '123456'
        with self.assertRaises(KeyError):
            state1[key]
        with self.assertRaises(KeyError):
            state2[key]
        state1[key] = value
        self.assertEqual(state1[key], value)
        self.assertEqual(state2[key], value)
        # Property
        state1.state_file = None
        state2.state_file = None
        self.assertEqual(state1.state_file, None)
        self.assertEqual(state2.state_file, None)
        value = '/some/path'
        state1.state_file = value
        self.assertEqual(state2.state_file, value)

    def test_unknown_attrib(self):
        with self.assertRaises(AttributeError):
            print(STATE.abcdef)

    def test_write(self):
        self.assertEqual(STATE.state_file, None)
        STATE.state_file = tempfile.mkstemp(prefix='vchtest')[1]
        STATE.write()
        with open(STATE.state_file, 'rb') as f:
            json.loads(f.read())

    def test_write_full(self):
        self.assertEqual(STATE.state_file, None)
        STATE.state_file = tempfile.mkstemp(prefix='vchtest')[1]
        STATE['disks'] = [Disk('/some/path'), Disk('/some/other/path', 12.34)]
        STATE['last_message'] = b'Byte data being saved'
        STATE.write()
        with open(STATE.state_file, 'rb') as f:
            json.loads(f.read())
