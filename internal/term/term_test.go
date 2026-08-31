package term_test

import (
	"os"
	"testing"

	"github.com/ironpark/runby/internal/term"
)

// TestNonTerminalsAreNotTerminals locks in the reason this package cannot be
// replaced by an os.ModeCharDevice test: /dev/null is a character device but
// is not a terminal, and reporting it as one would make every piped run look
// interactive.
func TestNonTerminalsAreNotTerminals(t *testing.T) {
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer devnull.Close()

	file, err := os.CreateTemp(t.TempDir(), "term")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer file.Close()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	defer reader.Close()
	defer writer.Close()

	for _, test := range []struct {
		name string
		fd   int
	}{
		{os.DevNull, int(devnull.Fd())},
		{"regular file", int(file.Fd())},
		{"pipe read end", int(reader.Fd())},
		{"pipe write end", int(writer.Fd())},
		{"unopened descriptor", 9999},
		{"negative descriptor", -1},
	} {
		if term.IsTerminal(test.fd) {
			t.Errorf("IsTerminal(%s) = true, want false", test.name)
		}
	}
}

// TestControllingTerminalIsATerminal covers the true case. It needs a
// controlling terminal, which a CI or piped run does not have.
func TestControllingTerminalIsATerminal(t *testing.T) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("no controlling terminal: %v", err)
	}
	defer tty.Close()

	if !term.IsTerminal(int(tty.Fd())) {
		t.Error("IsTerminal(/dev/tty) = false, want true")
	}
}
