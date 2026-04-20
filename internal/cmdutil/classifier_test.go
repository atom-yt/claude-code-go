package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassify_SafeCommands(t *testing.T) {
	tests := []struct {
		command string
		wantCat CommandCategory
		wantReason string
		wantReject bool
	}{
		{"echo hello", CategorySafe, "", false},
		{"ls -la", CategorySafe, "", false},
		{"cat file.txt", CategorySafe, "", false},
		{"grep pattern file", CategorySafe, "", false},
		{"go test ./...", CategorySafe, "", false},
		{"npm install", CategorySafe, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cat, reason, reject := Classify(tt.command)
			assert.Equal(t, tt.wantCat, cat)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantReject, reject)
		})
	}
}

func TestClassify_DestructiveCommands(t *testing.T) {
	tests := []struct {
		command string
		wantCat CommandCategory
		wantReason string
		wantReject bool
	}{
		{"rm -rf /", CategoryDestructive, "Recursive deletion from root", true},
		{"rm -rf /*", CategoryDestructive, "Recursive deletion from root", true},
		{"rm -rf ../", CategoryDestructive, "Recursive deletion to parent", true},
		{"rm /*", CategoryDestructive, "Deleting all files in root", true},
		{"rm -f /*", CategoryDestructive, "Force delete all in root", true},
		{"mkfs.ext4 /dev/sda1", CategoryDestructive, "Formatting filesystem", true},
		{"dd if=/dev/zero of=file", CategoryDestructive, "Low-level disk write", true},
		{"wipefs -a /dev/sda", CategoryDestructive, "Wiping filesystem signatures", true},
		{"shutdown now", CategoryDestructive, "System shutdown", true},
		{"reboot", CategoryDestructive, "System reboot", true},
		{"halt", CategoryDestructive, "System halt", true},
		{"poweroff", CategoryDestructive, "System poweroff", true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cat, reason, reject := Classify(tt.command)
			assert.Equal(t, tt.wantCat, cat, "command: %s", tt.command)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantReject, reject)
		})
	}
}

func TestClassify_PackageCommands(t *testing.T) {
	tests := []struct {
		command string
		wantCat CommandCategory
		wantReason string
		wantReject bool
	}{
		{"apt-get autoremove", CategoryPackage, "Removing packages", false},
		{"apt-get purge nginx", CategoryPackage, "Removing packages", false},
		{"apt remove package", CategoryPackage, "Removing packages", false},
		{"yum remove package", CategoryPackage, "Removing packages", false},
		{"dnf remove package", CategoryPackage, "Removing packages", false},
		{"pip uninstall requests", CategoryPackage, "Removing Python packages", false},
		{"npm uninstall -g package", CategoryPackage, "Removing global npm packages", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cat, reason, reject := Classify(tt.command)
			assert.Equal(t, tt.wantCat, cat)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantReject, reject)
		})
	}
}

func TestClassify_FileSystemCommands(t *testing.T) {
	tests := []struct {
		command string
		wantCat CommandCategory
		wantReason string
		wantReject bool
	}{
		{"chmod -R 777 /etc", CategoryFileSystem, "Recursive permission change", false},
		{"chown -R user:group /data", CategoryFileSystem, "Recursive ownership change", false},
		{"git branch -D feature", CategoryFileSystem, "Force deleting git branch", false},
		{"git push origin main --force", CategoryFileSystem, "Force pushing to git", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cat, reason, reject := Classify(tt.command)
			assert.Equal(t, tt.wantCat, cat)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantReject, reject)
		})
	}
}

func TestClassify_NetworkCommands(t *testing.T) {
	tests := []struct {
		command string
		wantCat CommandCategory
		wantReason string
		wantReject bool
	}{
		{"iptables -F", CategoryNetwork, "Firewall modification", false},
		{"ufw disable", CategoryNetwork, "Firewall disable/reset", false},
		{"ufw reset", CategoryNetwork, "Firewall disable/reset", false},
		{"nft list ruleset", CategoryNetwork, "Firewall modification", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cat, reason, reject := Classify(tt.command)
			assert.Equal(t, tt.wantCat, cat)
			assert.Equal(t, tt.wantReason, reason)
			assert.Equal(t, tt.wantReject, reject)
		})
	}
}

func TestIsDestructive(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"ls -la", false},
		{"cat file.txt", false},
		{"rm -rf /", true},
		{"rm -rf ..", true},
		{"dd if=/dev/zero of=file", true},
		{"mkfs.ext4 /dev/sda1", true},
		{"shutdown now", true},
		{"apt-get remove", false}, // package removal, not destructive
		{"chmod -R 777 /", false}, // filesystem, not destructive
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsDestructive(tt.command)
			assert.Equal(t, tt.want, result, "command: %s", tt.command)
		})
	}
}

func TestGetWarning(t *testing.T) {
	tests := []struct {
		command string
		wantEmpty bool
	}{
		{"echo hello", true},
		{"rm -rf /", false},
		{"apt-get remove", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := GetWarning(tt.command)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "⚠️")
			}
		})
	}
}