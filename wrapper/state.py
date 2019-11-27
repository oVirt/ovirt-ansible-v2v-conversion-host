import tempfile
import os
import json
import stat


class _StateObject(object):
    def as_dict(self):
        hidden = ['internal'] + getattr(self, '_hidden', [])
        slots = [key for key in getattr(self, '__slots__', self.__dict__)
                 if key not in hidden and not key.startswith('_')]
        return {key: getattr(self, key) for key in slots}


class _StateEncoder(json.JSONEncoder):
    def default(self, obj):  # pylint: disable=method-hidden
        if isinstance(obj, bytes):
            return obj.decode()
        if isinstance(obj, _StateObject):
            return obj.as_dict()
        return json.JSONEncoder.default(self, obj)


class Disk(_StateObject):
    """
    Represents one disk instance (to be) copied.
    """

    __slots__ = [
        'path',
        'progress',
    ]

    def __init__(self, path, progress=0.0):
        self.path = path
        self.progress = progress

    def __repr__(self):
        return "Disk(path=%s, progress=%.2f)" % (self.path, self.progress)


class _State(_StateObject):
    """
    State object using a dict for data storage.

    This is not just the contain of state file, but it contains all the
    internal configuration.

    Gradual conversion from the dict to properties is to be expected.
    """

    __slots__ = [
        'daemonize',
        'state_file',
        'v2v_log',
        'machine_readable_log',

        # These fields are written to the state file
        'disk_count',
        'disks',
        'failed',
        'finished',
        'internal',
        'last_message',
        'pid',
        'return_code',
        'started',
        'throttling',
        'vm_id',
    ]

    _hidden = [
        'daemonize',
        'state_file',
        'v2v_log',
        'machine_readable_log',
    ]

    def __init__(self):
        self.reset()

    def reset(self):
        """
        This function exists only so that the state can be re-used in tests
        """

        self.daemonize = True
        self.state_file = None
        self.v2v_log = None
        self.machine_readable_log = None

        self.disk_count = 0
        self.disks = []
        self.failed = False
        self.finished = None
        self.internal = {
            'disk_ids': {},
            'display_name': None,
            'ports': [],
            'throttling_file': None,
        }
        self.last_message = None
        self.pid = None
        self.started = None
        self.return_code = None
        # Does it make sense to create a throttling class?
        self.throttling = {'cpu': None, 'network': None}
        self.vm_id = None

    def write(self):
        tmp_state = tempfile.mkstemp(suffix='.v2v.state',
                                     dir=os.path.dirname(self.state_file))
        os.fchmod(tmp_state[0],
                  stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH)
        with os.fdopen(tmp_state[0], 'w') as f:
            json.dump(self, f, cls=_StateEncoder)
            os.rename(tmp_state[1], self.state_file)


STATE = _State()
