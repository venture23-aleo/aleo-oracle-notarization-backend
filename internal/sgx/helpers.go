package sgx

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// splitPath splits "a/b/c" into ["a","b","c"]
func splitPath(p string) []string {
	// Clean the path first
	p = filepath.Clean(p)

	// Split by "/" (Unix-style)
	parts := strings.Split(p, string(filepath.Separator))

	// Remove empty elements (e.g., leading "/")
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func SecureOpenFile(root, path string, mode string) (int, error) {

	// Validate mode
	if mode != ModeRead && mode != ModeWrite {
		return 0, fmt.Errorf("invalid mode: %s", mode)
	}

	cleanRoot := filepath.Clean(root)
	cleanPath := filepath.Clean(path)

	if strings.HasPrefix(cleanPath, cleanRoot) {
		rel, err := filepath.Rel(cleanRoot, cleanPath)
		if err == nil {
			path = rel
		}
	}

	// Split the path into segments
	segments := splitPath(path)

	// Validate segments
	if len(segments) == 0 {
		return 0, errors.New("invalid path")
	}

	rootFD, _ := syscall.Open(root, syscall.O_DIRECTORY, 0)
	defer syscall.Close(rootFD)

	// Start from root directory
	dirFD := rootFD

	// Get root device
	var rootStat syscall.Stat_t
	if err := syscall.Fstat(dirFD, &rootStat); err != nil {
		return 0, err
	}

	// Walk each path segment except last
	for i := 0; i < len(segments)-1; i++ {
		nextFd, err := syscall.Openat(dirFD, segments[i], syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
		if err != nil {
			return 0, err
		}

		var st syscall.Stat_t
		if err := syscall.Fstat(nextFd, &st); err != nil {
			syscall.Close(nextFd)
			return 0, err
		}

		// Must be a directory
		if (st.Mode & syscall.S_IFMT) != syscall.S_IFDIR {
			syscall.Close(nextFd)
			return 0, fmt.Errorf("%s is not a directory", segments[i])
		}

		// No symlinks allowed
		if (st.Mode & syscall.S_IFMT) == syscall.S_IFLNK {
			syscall.Close(nextFd)
			return 0, fmt.Errorf("symlink detected: %s", segments[i])
		}

		if st.Dev != rootStat.Dev {
			syscall.Close(nextFd)
			return 0, errors.New("cross-mount detected")
		}

		if dirFD != rootFD {
			syscall.Close(dirFD)
		}

		dirFD = nextFd
	}

	leafFlags := syscall.O_NOFOLLOW

	switch mode {
	case ModeRead:
		leafFlags |= syscall.O_RDONLY
	case ModeWrite:
		leafFlags |= syscall.O_WRONLY
	}

	leafFd, err := syscall.Openat(dirFD, segments[len(segments)-1], leafFlags, 0)

	if dirFD != rootFD {
		syscall.Close(dirFD)
	}

	return leafFd, err
}

func SecureReadFile(root, path string) ([]byte, error) {

	fd, err := SecureOpenFile(root, path, ModeRead)

	if err != nil {
		return nil, err
	}

	file := os.NewFile(uintptr(fd), path)
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func SecureWriteFile(root, path string, data []byte) error {

	fd, err := SecureOpenFile(root, path, ModeWrite)
	if err != nil {
		return err
	}

	file := os.NewFile(uintptr(fd), path)
	defer file.Close()

	_, err = file.Write(data)
	return err
}
