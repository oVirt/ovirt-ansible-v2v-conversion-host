import tempfile
import os
import json
import stat


class _StateObject(object):
    pass


class _StateEncoder(json.JSONEncoder):
    def default(self, obj):  # pylint: disable=method-hidden
        if isinstance(obj, bytes):
            return obj.decode()
        if isinstance(obj, _StateObject):
            hidden = ['internal']
            if hasattr(obj, '_hidden'):
                hidden += obj._hidden
            slots = [key for key in obj.__slots__ if key not in hidden]
            return {key: getattr(obj, key) for key in slots}
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


class _State(object):
    """
    State object using a dict for data storage.

    This is not just the contain of state file, but it contains all the
    internal configuration.

    Gradual conversion from the dict to properties is to be expected.
    """

    __slots__ = [
        '_state',  # Should be removed later

        'daemonize',
        'state_file',
        'v2v_log',
        'machine_readable_log',

        # These fields are written to the state file
        'failed',
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
        This function exists only so that the singleton can be re-used in tests
        """

        # For now keep content as dict. Idealy this should be changed
        # later too.
        self._state = {
            'disks': [],
            'internal': {
                'disk_ids': {},
                'display_name': None,
                'ports': [],
                'throttling_file': None,
            },
            'throttling': {
                'cpu': None,
                'network': None,
            }
        }
        self.daemonize = True
        self.state_file = None
        self.v2v_log = None
        self.machine_readable_log = None
        self.failed = False

    def __getattr__(self, name):
        return getattr(self._state, name)

    def __getitem__(self, key):
        return self._state[key]

    def __setitem__(self, key, value):
        self._state[key] = value

    def __str__(self):
        return repr(self._state)

    def write(self):
        hidden = ['internal']

        # Ideally this shenanigans will go away after all of the dict is
        # converted as we should be then able to just json.dumps(self).
        state = self._state.copy()
        for key in hidden:
            del state[key]

        if hasattr(self, '_hidden'):
            hidden += self._hidden
        slots = [key for key in self.__slots__
                 if key not in hidden and not key.startswith('_')]
        state.update({key: getattr(self, key) for key in slots})

        tmp_state = tempfile.mkstemp(suffix='.v2v.state',
                                     dir=os.path.dirname(self.state_file))
        os.fchmod(tmp_state[0],
                  stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH)
        with os.fdopen(tmp_state[0], 'w') as f:
            json.dump(state, f, cls=_StateEncoder)
            os.rename(tmp_state[1], self.state_file)


STATE = _State()
