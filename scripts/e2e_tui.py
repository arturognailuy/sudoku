#!/usr/bin/env python3
"""Black-box pseudo-terminal smoke test for `sudoku tui`."""
import argparse
import fcntl
import os
import pty
import re
import select
import signal
import struct
import subprocess
import tempfile
import termios
import time

PUZZLE = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
ANSI = re.compile(rb"\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07]*(?:\x07|\x1b\\)|[()][A-Z0-9])")


def drain(fd, seconds=0.25):
    end = time.monotonic() + seconds
    chunks = []
    while time.monotonic() < end:
        ready, _, _ = select.select([fd], [], [], 0.05)
        if not ready:
            continue
        try:
            chunks.append(os.read(fd, 65536))
        except OSError:
            break
    return b"".join(chunks)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("binary", nargs="?", default="./sudoku")
    args = parser.parse_args()
    with tempfile.TemporaryDirectory() as directory:
        session = os.path.join(directory, "tui-session.json")
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 42, 90, 0, 0))
        env = os.environ.copy()
        env["XDG_DATA_HOME"] = os.path.join(directory, "data")
        process = subprocess.Popen(
            [args.binary, "tui", "--input", PUZZLE],
            stdin=slave, stdout=slave, stderr=slave, env=env, close_fds=True,
        )
        os.close(slave)
        output = drain(master, 0.8)
        # Resize below and back above the minimum; game state must survive.
        fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 20, 40, 0, 0))
        process.send_signal(signal.SIGWINCH)
        small_output = drain(master, 0.4)
        if b"Terminal too small" not in ANSI.sub(b"", small_output):
            raise AssertionError("small-terminal fallback was not rendered")
        output += small_output
        fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 42, 90, 0, 0))
        process.send_signal(signal.SIGWINCH)
        output += drain(master, 0.4)
        # Move to an editable cell, set a value, toggle a note, exercise history
        # and hint preview/apply, then save explicitly and cleanly quit.
        for keys in (b"l", b"5", b"q", b"n", b"n", b"j", b"4", b"u", b"r", b"?", b"\r", b"S"):
            os.write(master, keys)
            output += drain(master)
        os.write(master, session.encode() + b"\r")
        output += drain(master, 0.5)
        os.write(master, b"q")
        output += drain(master, 0.5)
        try:
            return_code = process.wait(timeout=3)
        except subprocess.TimeoutExpired:
            process.kill()
            raise AssertionError("TUI did not quit after a clean save")
        finally:
            os.close(master)
        text = ANSI.sub(b"", output).decode("utf-8", "replace")
        required = ("Sudoku", "mode:NOTE", "Hint preview:", "Saved to ", "unsaved:*", "Unsaved changes")
        missing = [value for value in required if value not in text]
        if missing:
            raise AssertionError(f"screen output missing {missing}\n{text[-4000:]}")
        if return_code != 0:
            raise AssertionError(f"TUI exited {return_code}\n{text[-2000:]}")
        if not os.path.isfile(session) or os.stat(session).st_mode & 0o777 != 0o600:
            raise AssertionError("explicit save did not create a mode-0600 session")
        # A saved session must start in a second full-screen process and quit cleanly.
        master, slave = pty.openpty()
        fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 42, 90, 0, 0))
        resumed = subprocess.Popen([args.binary, "tui", "--resume", session], stdin=slave, stdout=slave, stderr=slave, env=env, close_fds=True)
        os.close(slave)
        resumed_output = drain(master, 0.7)
        os.write(master, b"q")
        resumed_output += drain(master, 0.3)
        if resumed.wait(timeout=3) != 0 or b"Sudoku" not in ANSI.sub(b"", resumed_output):
            raise AssertionError("saved TUI session did not resume cleanly")
        os.close(master)
        corrupt = os.path.join(directory, "corrupt.json")
        with open(corrupt, "w", encoding="utf-8") as handle:
            handle.write("{bad json")
        rejected = subprocess.run([args.binary, "tui", "--resume", corrupt], stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=env, timeout=3)
        if rejected.returncode == 0 or b"resume saved session" not in rejected.stdout.lower():
            raise AssertionError("corrupt TUI restore was not rejected before startup")
        print("PASS: TUI PTY resize, input, confirmation, history, hint, save/resume, and quit")


if __name__ == "__main__":
    main()
