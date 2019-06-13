package main

import (
	"errors"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
	"unsafe"
)

func main() {
	Notify("/tmp/git/link", func(ev *Event) { log.Println("event:", ev.Name) })
}

// Notify calls fn whenever link is moved to (as per inotify).  This will not
// return unless there is an error.
func Notify(link string, fn func(*Event)) error {
	n, err := newNotifier()
	if err != nil {
		return err
	}
	dir, file := path.Split(link)
	err = n.watch(dir)
	if err != nil {
		return err
	}
	for {
		select {
		case ev := <-n.Event:
			if ev.Name == file {
				fn(ev)
			}
		case err := <-n.Error:
			return err
		}
	}
	// Should never get here.
	return nil
}

type Event struct {
	Mask   uint32 // Mask of events
	Cookie uint32 // Unique cookie associating related events (for rename(2))
	Name   string // File name (optional)
}

type notifier struct {
	fd    int         // File descriptor (as returned by the inotify_init() syscall)
	wd    uint32      // Watch descriptor (as returned by the inotify_add_watch() syscall)
	Error chan error  // Errors are sent on this channel
	Event chan *Event // Events are returned on this channel
}

// newNotifier creates and returns a new inotify instance using inotify_init(2)
func newNotifier() (*notifier, error) {
	fd, errno := syscall.InotifyInit()
	if fd == -1 {
		return nil, os.NewSyscallError("inotify_init", errno)
	}
	notifier := &notifier{
		fd:    fd,
		Event: make(chan *Event),
		Error: make(chan error),
	}

	go notifier.readEvents()
	return notifier, nil
}

// watch watches dir for IN_MOVED_TO events.
func (notifier *notifier) watch(dir string) error {
	wd, err := syscall.InotifyAddWatch(notifier.fd, dir, IN_MOVED_TO|IN_DONT_FOLLOW|IN_ONLYDIR)
	if err != nil {
		return &os.PathError{
			Op:   "inotify_add_watch",
			Path: dir,
			Err:  err,
		}
	}

	notifier.wd = uint32(wd)
	return nil
}

// readEvents reads from the inotify file descriptor, converts the
// received events into Event objects and sends them via the Event channel
func (notifier *notifier) readEvents() {
	var buf [syscall.SizeofInotifyEvent * 4096]byte

	for {
		n, err := syscall.Read(notifier.fd, buf[:])

		// If EOF...
		if n == 0 {
			// The syscall.Close can be slow.  Close
			// notifier.Event first.
			close(notifier.Event)
			err := syscall.Close(notifier.fd)
			if err != nil {
				notifier.Error <- os.NewSyscallError("close", err)
			}
			close(notifier.Error)
			return
		}
		if n < 0 {
			notifier.Error <- os.NewSyscallError("read", err)
			continue
		}
		if n < syscall.SizeofInotifyEvent {
			notifier.Error <- errors.New("inotify: short read in readEvents()")
			continue
		}

		var offset uint32 = 0
		// We don't know how many events we just read into the buffer
		// While the offset points to at least one whole event...
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			// Point "raw" to the event in the buffer
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			event := new(Event)
			event.Mask = uint32(raw.Mask)
			event.Cookie = uint32(raw.Cookie)
			nameLen := uint32(raw.Len)
			if nameLen > 0 {
				// Point "bytes" at the first byte of the filename
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				// The filename is padded with NUL bytes. TrimRight() gets rid of those.
				event.Name = strings.TrimRight(string(bytes[0:nameLen]), "\000")
			}
			// Send the event on the events channel
			notifier.Event <- event
			// Move to the next event in the buffer
			offset += syscall.SizeofInotifyEvent + nameLen
		}
	}
}

const (
	// Options for AddWatch
	IN_DONT_FOLLOW uint32 = syscall.IN_DONT_FOLLOW
	IN_ONLYDIR     uint32 = syscall.IN_ONLYDIR

	// Events
	IN_MOVED_TO uint32 = syscall.IN_MOVED_TO
)
