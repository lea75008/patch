Replace `unix.Chmod()` with `unix.Fchmod()` using a file descriptor opened with `O_NOFOLLOW`:

```go
func (m *mknodUnix) Mknode(path string, major, minor int) error {
    if _, err := os.Stat(path); err == nil {
        m.logger.Infof("Skipping: %s already exists", path)
        return nil
    } else if !os.IsNotExist(err) {
        return fmt.Errorf("failed to stat %s: %v", path, err)
    }

    err := unix.Mknod(path, unix.S_IFCHR, int(unix.Mkdev(uint32(major), uint32(minor))))
    if err != nil {
        return err
    }

    // FIXED: Open with O_NOFOLLOW to prevent symlink following
    fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
    if err != nil {
        if err == unix.ELOOP {
            // Path is a symlink - this shouldn't happen, fail safely
            unix.Unlink(path)
            return fmt.Errorf("unexpected symlink at device path %s", path)
        }
        return fmt.Errorf("failed to open device node: %w", err)
    }
    defer unix.Close(fd)

    // Fchmod operates on the fd, not the path - cannot follow symlinks
    return unix.Fchmod(fd, 0666)
}
