package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/youyo/imgraft/internal/auth"
)

// TestLogout_RemovesGoogleAIStudio は google_ai_studio 認証の削除をテスト。
func TestLogout_RemovesGoogleAIStudio(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	// credentials を事前に保存
	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key-to-remove"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	opts := auth.LogoutOptions{
		Profile:  "default",
		Backend:  "google_ai_studio",
		CredPath: credPath,
	}

	if err := auth.Logout(opts); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// 削除後の確認
	loaded, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	pc := loaded.Profiles["default"]
	if pc.GoogleAIStudio != nil {
		t.Error("GoogleAIStudio should be nil after logout")
	}
}

// TestLogout_ProfileNotDeleted はログアウト後も profile 自体は残ることをテスト。
func TestLogout_ProfileNotDeleted(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	opts := auth.LogoutOptions{
		Profile:  "default",
		Backend:  "google_ai_studio",
		CredPath: credPath,
	}

	if err := auth.Logout(opts); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	loaded, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// profile 自体は残っている
	if _, ok := loaded.Profiles["default"]; !ok {
		t.Error("profile 'default' should still exist after logout")
	}
}

// TestLogout_NoCredentials は credentials.json がない場合もエラーにならないことをテスト。
func TestLogout_NoCredentials(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "nonexistent.json")

	opts := auth.LogoutOptions{
		Profile:  "default",
		Backend:  "google_ai_studio",
		CredPath: credPath,
	}

	// ファイルが存在しない場合はエラーなし（ログアウト済み扱い）
	if err := auth.Logout(opts); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
}

// TestLogout_MultipleProfiles は複数 profile 環境でのログアウトをテスト。
func TestLogout_MultipleProfiles(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key-default"},
			},
			"work": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key-work"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	opts := auth.LogoutOptions{
		Profile:  "default",
		Backend:  "google_ai_studio",
		CredPath: credPath,
	}

	if err := auth.Logout(opts); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	loaded, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// default はログアウト済み
	if loaded.Profiles["default"].GoogleAIStudio != nil {
		t.Error("default GoogleAIStudio should be nil after logout")
	}
	// work は影響なし
	if loaded.Profiles["work"].GoogleAIStudio == nil || loaded.Profiles["work"].GoogleAIStudio.APIKey != "key-work" {
		t.Error("work profile should be unchanged")
	}
}

// TestLogout_DefaultBackend は backend 未指定時に google_ai_studio を使うことをテスト。
func TestLogout_DefaultBackend(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	opts := auth.LogoutOptions{
		Profile:  "default",
		Backend:  "", // 空文字 = デフォルト
		CredPath: credPath,
	}

	if err := auth.Logout(opts); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	loaded, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !os.IsNotExist(nil) { // 常に true だが存在チェックのダミー
		pc := loaded.Profiles["default"]
		if pc.GoogleAIStudio != nil {
			t.Error("GoogleAIStudio should be nil after logout with default backend")
		}
	}
}
