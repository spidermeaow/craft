package project

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func optionalLock(root string) (projectLock, error) {
	p := filepath.Join(root, lockFileName)
	if _, e := os.Lstat(p); os.IsNotExist(e) {
		return projectLock{}, nil
	}
	return readProjectLock(p)
}

const installJournal = ".craft-install-journal.json"

func checkInstallJournal(root string) error {
	if _, e := os.Lstat(filepath.Join(root, installJournal)); e == nil {
		return fmt.Errorf("interrupted package transaction; run craft install --recover before using the project")
	} else if !os.IsNotExist(e) {
		return e
	}
	return nil
}

type InstallOptions struct {
	Spec, Alias               string
	Offline, Refresh, Recover bool
}

// Install only fetches source. It never runs dependency tests, builds or hooks.
func Install(ctx context.Context, start string, o InstallOptions) ([]string, error) {
	return installWithTransport(ctx, start, o, nil)
}
func installWithTransport(ctx context.Context, start string, o InstallOptions, transport packageTransport) ([]string, error) {
	root, e := Find(start)
	if e != nil {
		return nil, e
	}
	unlock, e := lockInstallation(root)
	if e != nil {
		return nil, e
	}
	defer unlock()
	if o.Recover {
		if o.Spec != "" || o.Alias != "" || o.Refresh || o.Offline {
			return nil, fmt.Errorf("--recover must be used alone")
		}
		if e = recoverInstallation(root); e != nil {
			return nil, e
		}
		return []string{"Recovered package transaction; run craft install to restore."}, nil
	}
	if e = checkInstallJournal(root); e != nil {
		return nil, e
	}
	original, e := readLimited(filepath.Join(root, "craft.toml"), 64<<10)
	if e != nil {
		return nil, e
	}
	p, e := load(root, false)
	if e != nil {
		return nil, e
	}
	lock, e := optionalLock(root)
	if e != nil {
		return nil, e
	}
	oldLockBytes, e := os.ReadFile(filepath.Join(root, lockFileName))
	if e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	r, e := newRemoteResolver(ctx, lock, !o.Offline)
	if e != nil {
		return nil, e
	}
	if transport != nil {
		r.transport = transport
	}
	if o.Refresh && (o.Spec == "" || o.Offline) {
		return nil, fmt.Errorf("--refresh requires an explicit source and online access")
	}
	if o.Alias != "" && o.Spec == "" {
		return nil, fmt.Errorf("--alias requires an explicit source")
	}
	updated := original
	messages := []string{}
	if o.Spec != "" {
		d, err := parseGitHubSpec(o.Spec)
		if err != nil {
			return nil, err
		}
		previous, hadPin := r.pins[d.key()]
		if o.Refresh {
			delete(r.pins, d.key())
		}
		dir, pin, err := r.ensure(d)
		if err != nil {
			return nil, err
		}
		m, err := ReadManifest(dir)
		if err != nil {
			return nil, err
		}
		alias := o.Alias
		if alias == "" {
			alias = strings.ReplaceAll(m.Name, "-", "_")
		}
		if !validAlias(alias) {
			return nil, fmt.Errorf("invalid alias; choose --alias <identifier>")
		}
		if _, exists := p.Manifest.Dependencies[alias]; exists {
			old, remote := p.Manifest.GitHub[alias]
			if !remote || old.Repository != d.Repository {
				return nil, fmt.Errorf("alias %s already refers to another source; choose --alias", alias)
			}
		}
		updated, err = editGitHubDependency(original, p.Manifest, alias, d)
		if err != nil {
			return nil, err
		}
		p.Manifest, err = readManifestData(root, updated)
		if err != nil {
			return nil, err
		}
		messages = append(messages, fmt.Sprintf("Selected %s as %s, commit %s, sha256:%s", d.key(), alias, pin.Commit, pin.Checksum))
		if hadPin && previous.Commit != pin.Commit {
			messages = append(messages, fmt.Sprintf("Explicit refresh changed commit %s -> %s", previous.Commit, pin.Commit))
		}
	}
	// Clear the visited set populated by the alias probe: every pin must be used by the graph.
	for k, pin := range r.used {
		r.pins[k] = pin
	}
	r.used = map[string]RemotePin{}
	r.checked = map[string]string{}
	p, e = resolveGraph(p, false, r)
	if e != nil {
		return nil, e
	}
	if _, _, e = p.CompileMode(p.Manifest.Entry != "library"); e != nil {
		return nil, e
	}
	next := currentLock(p)
	if o.Spec == "" && lock.Format != 0 {
		a, _ := json.Marshal(lock)
		b, _ := json.Marshal(next)
		if !bytes.Equal(a, b) {
			return nil, fmt.Errorf("manifest/source does not match craft.lock; review local changes with craft package lock, or use explicit craft install <source@tag>")
		}
	}
	data, e := json.MarshalIndent(next, "", "  ")
	if e != nil {
		return nil, e
	}
	data = append(data, '\n')
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	latestLock, e := os.ReadFile(filepath.Join(root, lockFileName))
	if e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	if !bytes.Equal(oldLockBytes, latestLock) {
		return nil, fmt.Errorf("lock changed during install; retry")
	}
	if e = commitInstallation(root, original, updated, data); e != nil {
		return nil, e
	}
	messages = append(messages, fmt.Sprintf("Installed %d dependencies; manifest and lock are synchronized.", len(p.Packages)))
	return messages, nil
}

func editGitHubDependency(data []byte, m Manifest, alias string, d GitHubDependency) ([]byte, error) {
	line := fmt.Sprintf("%s = { github = %q, tag = %q, version = %q }", alias, d.Repository, d.Tag, d.Version)
	text := string(data)
	newline := "\n"
	if strings.Contains(text, "\r\n") {
		newline = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if pos, ok := m.DependencyPositions[alias]; ok {
		idx := pos.Line - 1
		if idx < 0 || idx >= len(lines) {
			return nil, fmt.Errorf("invalid manifest position")
		}
		// Keep trailing comments without interpreting # inside a quoted value.
		quote := rune(0)
		escaped := false
		comment := ""
		for i, c := range lines[idx] {
			if quote != 0 {
				if escaped {
					escaped = false
					continue
				}
				if quote == '"' && c == '\\' {
					escaped = true
					continue
				}
				if c == quote {
					quote = 0
				}
				continue
			}
			if c == '\'' || c == '"' {
				quote = c
			}
			if c == '#' {
				comment = " " + lines[idx][i:]
				break
			}
		}
		lines[idx] = line + comment
	} else {
		found := false
		for _, v := range lines {
			if strings.TrimSpace(strings.SplitN(v, "#", 2)[0]) == "[dependencies]" {
				found = true
			}
		}
		if !found {
			lines = append(lines, "[dependencies]")
		}
		lines = append(lines, line)
	}
	result := []byte(strings.TrimRight(strings.Join(lines, newline), "\r\n") + newline)
	if len(result) > 64<<10 {
		return nil, fmt.Errorf("manifest exceeds limit")
	}
	return result, nil
}

type installTransaction struct {
	OldManifest, NewManifest, OldLock, NewLock []byte
	HadLock                                    bool
}

func replacePackageFile(path string, data []byte) error {
	if info, e := os.Lstat(path); e == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing non-regular package file")
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".craft-write-")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if _, e = f.Write(data); e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}
func commitInstallation(root string, oldManifest, newManifest, newLock []byte) error {
	return commitInstallationWith(root, oldManifest, newManifest, newLock, replacePackageFile)
}
func commitInstallationWith(root string, oldManifest, newManifest, newLock []byte, replace func(string, []byte) error) error {
	oldLock, e := os.ReadFile(filepath.Join(root, lockFileName))
	had := e == nil
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	actual, e := os.ReadFile(filepath.Join(root, "craft.toml"))
	if e != nil || !bytes.Equal(actual, oldManifest) {
		return fmt.Errorf("manifest changed during install; retry")
	}
	tx := installTransaction{oldManifest, newManifest, oldLock, newLock, had}
	data, e := json.Marshal(tx)
	if e != nil {
		return e
	}
	j, e := os.OpenFile(filepath.Join(root, installJournal), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = j.Write(data)
	if e == nil {
		e = j.Sync()
	}
	closeErr := j.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	// Once the journal exists, all project loads fail closed until completion/recovery.
	for _, item := range []struct {
		name string
		data []byte
	}{{"craft.toml", newManifest}, {lockFileName, newLock}} {
		if e = replace(filepath.Join(root, item.name), item.data); e != nil {
			recoveryErr := recoverInstallation(root)
			if recoveryErr != nil {
				return fmt.Errorf("install failed; recovery required: %v", recoveryErr)
			}
			return e
		}
	}
	return os.Remove(filepath.Join(root, installJournal))
}
func recoverInstallation(root string) error {
	data, e := readLimited(filepath.Join(root, installJournal), 4<<20)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	var tx installTransaction
	if json.Unmarshal(data, &tx) != nil || len(tx.OldManifest) == 0 {
		return fmt.Errorf("invalid install journal; manual review required")
	}
	if _, e = readManifestData(root, tx.OldManifest); e != nil {
		return fmt.Errorf("invalid original manifest in journal")
	}
	for _, item := range []struct {
		name      string
		old, next []byte
		existed   bool
	}{{"craft.toml", tx.OldManifest, tx.NewManifest, true}, {lockFileName, tx.OldLock, tx.NewLock, tx.HadLock}} {
		p := filepath.Join(root, item.name)
		actual, err := os.ReadFile(p)
		if err != nil && !(os.IsNotExist(err) && !item.existed) {
			return err
		}
		if !bytes.Equal(actual, item.old) && !bytes.Equal(actual, item.next) {
			return fmt.Errorf("%s was edited after interrupted install; preserve edits and recover manually", item.name)
		}
	}
	if e = replacePackageFile(filepath.Join(root, "craft.toml"), tx.OldManifest); e != nil {
		return e
	}
	if tx.HadLock {
		e = replacePackageFile(filepath.Join(root, lockFileName), tx.OldLock)
	} else {
		e = os.Remove(filepath.Join(root, lockFileName))
		if os.IsNotExist(e) {
			e = nil
		}
	}
	if e != nil {
		return e
	}
	return os.Remove(filepath.Join(root, installJournal))
}
