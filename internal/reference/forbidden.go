package reference

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/youyo/imgraft/internal/errs"
)

// privateRanges はプライベートネットワーク CIDR リスト。
// SPEC.md セクション 10.8 に準拠する。
var privateRanges []*net.IPNet

func init() {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",  // IPv6 unique local
		"fe80::/10", // IPv6 link-local
	}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			privateRanges = append(privateRanges, network)
		}
	}
}

// ValidateURL は rawURL のホストを検証し、プライベート IP / localhost の場合はエラーを返す。
// サポートするスキームは http:// と https:// のみ。
// SPEC.md セクション 10.8 を参照。
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return errs.New(errs.ErrReferenceURLForbidden, fmt.Sprintf("invalid URL: %v", err))
	}

	// スキームチェック: http/https のみ許可
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return errs.New(errs.ErrReferenceURLForbidden,
			fmt.Sprintf("unsupported URL scheme %q: only http and https are allowed", u.Scheme))
	}

	host := u.Hostname()
	if host == "" {
		return errs.New(errs.ErrReferenceURLForbidden, "URL has no host")
	}

	// localhost 拒否
	if strings.ToLower(host) == "localhost" {
		return errs.New(errs.ErrReferenceURLForbidden,
			fmt.Sprintf("access to %q is forbidden", host))
	}

	// IP アドレスとして解析してプライベートレンジを確認
	ip := net.ParseIP(host)
	if ip != nil {
		if isPrivateIP(ip) {
			return errs.New(errs.ErrReferenceURLForbidden,
				fmt.Sprintf("access to private/loopback IP %q is forbidden", host))
		}
	}

	return nil
}

// isPrivateIP は IP がプライベートレンジに属するか判定する。
func isPrivateIP(ip net.IP) bool {
	for _, network := range privateRanges {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
