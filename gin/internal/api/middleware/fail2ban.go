package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Firewall name
const FixedRuleName = "Golang_Block_0"

// ReadBlockedIPsFromNetsh queries Windows Firewall rule "Golang_Block_0" to parse existing blocked remote IPs.
//
// Returns:
//   - []string: slice of currently blocked remote IP addresses parsed from Windows Firewall.
func ReadBlockedIPsFromNetsh() []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name="+FixedRuleName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	var existingIPs []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "remoteip") || strings.Contains(lower, "ip jarak jauh") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				ipListStr := strings.TrimSpace(parts[1])
				if ipListStr == "Any" || ipListStr == "Semua" || ipListStr == "*" {
					continue
				}
				rawIPs := strings.Split(ipListStr, ",")
				for _, raw := range rawIPs {
					ip := strings.TrimSpace(raw)
					if idx := strings.Index(ip, "/"); idx != -1 {
						ip = ip[:idx]
					}
					if host, _, err := net.SplitHostPort(ip); err == nil {
						ip = host
					}
					if ip != "" && ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
						existingIPs = append(existingIPs, ip)
					}
				}
			}
		}
	}

	return existingIPs
}

// SyncBlockedIPs_Netsh updates or creates the single firewall rule "Golang_Block_0" with comma-separated unique IPs.
//
// Parameters:
//   - allIPs: slice of unique IP addresses to enforce in the firewall blocklist.
//
// Returns:
//   - error: non-nil if executing the netsh firewall command fails.
func SyncBlockedIPs_Netsh(allIPs []string) error {
	if len(allIPs) == 0 {
		return nil
	}

	remoteIPs := strings.Join(allIPs, ",")

	if runtime.GOOS != "windows" {
		log.Printf("[Fail2Ban] [NON-WINDOWS OS: %s] Simulated firewall rule '%s' updated with remoteip=%s\n", runtime.GOOS, FixedRuleName, remoteIPs)
		return nil
	}

	// 1. Try updating existing rule "Golang_Block_0"
	cmdSet := exec.Command("netsh", "advfirewall", "firewall", "set", "rule",
		"name="+FixedRuleName,
		"new",
		"remoteip="+remoteIPs)

	outputSet, errSet := cmdSet.CombinedOutput()
	outSetStr := string(outputSet)

	if errSet == nil && (strings.Contains(outSetStr, "Ok") || strings.Contains(outSetStr, "OK") || strings.Contains(outSetStr, "ok")) {
		log.Printf("[Fail2Ban] [WINDOWS FIREWALL] Updated rule '%s' with remoteip=%s\n", FixedRuleName, remoteIPs)
		return nil
	}

	// 2. If setting existing rule failed (rule doesn't exist yet), create it
	cmdAdd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name="+FixedRuleName,
		"dir=in",
		"action=block",
		"remoteip="+remoteIPs)

	outputAdd, errAdd := cmdAdd.CombinedOutput()
	if errAdd != nil {
		return fmt.Errorf("gagal membuat/mengedit firewall rule '%s': %v, output: %s", FixedRuleName, errAdd, string(outputAdd))
	}

	outAddStr := string(outputAdd)
	if !strings.Contains(outAddStr, "Ok") && !strings.Contains(outAddStr, "OK") && !strings.Contains(outAddStr, "ok") {
		return fmt.Errorf("firewall mereturn pesan tidak terduga saat membuat rule '%s': %s", FixedRuleName, outAddStr)
	}

	log.Printf("[Fail2Ban] [WINDOWS FIREWALL] Created rule '%s' with remoteip=%s\n", FixedRuleName, remoteIPs)
	return nil
}

// BlockIP_Netsh executes netsh command on Windows to add/update an IP in the "Golang_Block_0" firewall rule.
//
// Parameters:
//   - ip: remote IP address to block.
//
// Returns:
//   - error: non-nil if firewall command fails.
func BlockIPNetsh(ip string) error {
	return GetFail2Ban().blockIP(ip)
}

// IsScannerURL detects common vulnerability/scanner probe URL patterns (e.g. .env, wp-admin, .php).
//
// Parameters:
//   - path: request URL path string.
//
// Returns:
//   - bool: true if path matches known scanner or exploit probe patterns.
func IsScannerURL(path string) bool {
	p := strings.ToLower(strings.TrimSpace(path))

	// Sensitive file / scanner probes
	if strings.Contains(p, ".env") ||
		strings.Contains(p, ".git") ||
		strings.Contains(p, ".svn") ||
		strings.Contains(p, ".htaccess") ||
		strings.Contains(p, ".aws") ||
		strings.Contains(p, ".ssh") ||
		strings.Contains(p, "wp-admin") ||
		strings.Contains(p, "wp-login") ||
		strings.Contains(p, "wp-content") ||
		strings.Contains(p, "wp-includes") ||
		strings.Contains(p, "xmlrpc.php") ||
		strings.Contains(p, "phpmyadmin") ||
		strings.Contains(p, "myadmin") ||
		strings.Contains(p, "pma") ||
		strings.Contains(p, "actuator") ||
		strings.Contains(p, "eval-stdin") ||
		strings.Contains(p, "shell.php") ||
		strings.Contains(p, "setup.php") ||
		strings.Contains(p, "install.php") {
		return true
	}

	// Common scanner script extensions
	if strings.HasSuffix(p, ".php") ||
		strings.HasSuffix(p, ".asp") ||
		strings.HasSuffix(p, ".aspx") ||
		strings.HasSuffix(p, ".jsp") ||
		strings.HasSuffix(p, ".cgi") ||
		strings.HasSuffix(p, ".bak") ||
		strings.HasSuffix(p, ".sql") ||
		strings.HasSuffix(p, ".config") {
		return true
	}

	return false
}

// ipTracker records timestamps of general and vulnerability scanner 404 occurrences per IP.
type ipTracker struct {
	mu          sync.Mutex
	general404s []time.Time
	scanner404s []time.Time
}

// Fail2Ban provides automated intrusion detection and IP blocking integration with Windows Firewall.
type Fail2Ban struct {
	blockedIPs sync.Map // map[string]bool
	trackers   sync.Map // map[string]*ipTracker
	stopChan   chan struct{}
}

var (
	globalFail2Ban     *Fail2Ban
	globalFail2BanOnce sync.Once
)

// GetFail2Ban returns the singleton instance of Fail2Ban with background cleanup loop running.
//
// Returns:
//   - *Fail2Ban: singleton Fail2Ban instance.
func GetFail2Ban() *Fail2Ban {
	globalFail2BanOnce.Do(func() {
		globalFail2Ban = NewFail2Ban()
	})
	return globalFail2Ban
}

// NewFail2Ban instantiates and starts a new Fail2Ban manager with background tracker cleanup.
//
// Returns:
//   - *Fail2Ban: new Fail2Ban instance.
func NewFail2Ban() *Fail2Ban {
	fb := &Fail2Ban{
		stopChan: make(chan struct{}),
	}
	go fb.startCleanupLoop()
	return fb
}

// blockIP adds an IP to the in-memory block map and syncs it to the Windows Firewall rule.
//
// Parameters:
//   - ip: remote IP address to block.
//
// Returns:
//   - error: non-nil if firewall sync fails.
func (f *Fail2Ban) blockIP(ip string) error {
	ip = strings.TrimSpace(ip)
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	if ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return nil
	}

	f.blockedIPs.Store(ip, true)

	// Re-read existing blocked IPs from Windows Firewall to keep Go memory footprint low
	existingIPs := ReadBlockedIPsFromNetsh()

	// Merge existing IPs with new IP to create a unique set
	uniqueMap := make(map[string]bool)
	for _, ex := range existingIPs {
		uniqueMap[ex] = true
	}
	uniqueMap[ip] = true

	allIPs := make([]string, 0, len(uniqueMap))
	for k := range uniqueMap {
		allIPs = append(allIPs, k)
	}

	return SyncBlockedIPs_Netsh(allIPs)
}

// Stop gracefully terminates the background cleanup loop.
func (f *Fail2Ban) Stop() {
	select {
	case <-f.stopChan:
		// Already stopped
	default:
		close(f.stopChan)
	}
}

// IsBlocked checks whether the given IP is currently recorded in the active blocklist.
//
// Parameters:
//   - ip: remote IP address to test.
//
// Returns:
//   - bool: true if the IP is blocked.
func (f *Fail2Ban) IsBlocked(ip string) bool {
	ip = sanitizeIP(ip)
	_, blocked := f.blockedIPs.Load(ip)
	return blocked
}

// Record404 evaluates a 404 Not Found event against two threshold rules:
//   - General probe rule: 10 occurrences within 20 seconds.
//   - Scanner probe rule: 2 scanner URL probes within 60 seconds.
//
// Parameters:
//   - ip: client remote IP address.
//   - path: requested URL path.
//
// Returns:
//   - bool: true if IP exceeded threshold and was blocked.
func (f *Fail2Ban) Record404(ip, path string) bool {
	ip = sanitizeIP(ip)

	// Never block localhost / loopback or private network IPs
	parsedIP := net.ParseIP(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost" || (parsedIP != nil && (parsedIP.IsLoopback() || parsedIP.IsPrivate())) {
		return false
	}

	if f.IsBlocked(ip) {
		return true
	}

	now := time.Now()
	val, _ := f.trackers.LoadOrStore(ip, &ipTracker{})
	tracker := val.(*ipTracker)

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	shouldBlock := false
	reason := ""

	if IsScannerURL(path) {
		// Rule B: 2 scanner 404s within 60 seconds (1 minute)
		tracker.scanner404s = append(tracker.scanner404s, now)
		validScanner := make([]time.Time, 0, len(tracker.scanner404s))
		for _, t := range tracker.scanner404s {
			if now.Sub(t) <= 60*time.Second {
				validScanner = append(validScanner, t)
			}
		}
		tracker.scanner404s = validScanner

		if len(tracker.scanner404s) >= 2 {
			shouldBlock = true
			reason = fmt.Sprintf("2x 404 on Scanner URL probe within 1 min (%s)", path)
		}
	}

	// Always track general 404
	// Rule A: 10 general 404s within 20 seconds
	tracker.general404s = append(tracker.general404s, now)
	validGeneral := make([]time.Time, 0, len(tracker.general404s))
	for _, t := range tracker.general404s {
		if now.Sub(t) <= 20*time.Second {
			validGeneral = append(validGeneral, t)
		}
	}
	tracker.general404s = validGeneral

	if len(tracker.general404s) >= 10 {
		shouldBlock = true
		reason = "10x 404 within 20 seconds"
	}

	if shouldBlock {
		log.Printf("[Fail2Ban] BLOCKED IP %s | Reason: %s\n", ip, reason)
		if err := f.blockIP(ip); err != nil {
			log.Printf("[Fail2Ban] Error executing firewall block for %s: %v\n", ip, err)
		}
		return true
	}

	return false
}

// Middleware returns Gin middleware that checks IP block status and tracks 404 responses asynchronously.
//
// Returns:
//   - gin.HandlerFunc: intrusion detection middleware handler.
func (f *Fail2Ban) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := sanitizeIP(c.ClientIP())

		if f.IsBlocked(ip) {
			if gin.IsDebugging() {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access blocked by firewall rules",
					"ip":    ip,
				})
			} else {
				candidates := []string{
					"errors/404.htm",
					"./errors/404.htm",
					"dist/errors/404.htm",
					"gin/dist/errors/404.htm",
					"/app/errors/404.htm",
				}
				for _, path := range candidates {
					if data, err := os.ReadFile(path); err == nil {
						c.Data(http.StatusNotFound, "text/html; charset=utf-8", data)
						c.Abort()
						return
					}
				}
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Not Found",
				})
			}
			c.Abort()
			return
		}

		c.Next()

		if c.Writer.Status() == http.StatusNotFound {
			reqPath := c.Request.URL.Path
			// Fire-and-forget goroutine so response is not delayed
			go f.Record404(ip, reqPath)
		}
	}
}

// CleanupStaleTrackers removes expired 404 history entries and purges inactive IP trackers.
func (f *Fail2Ban) CleanupStaleTrackers() {
	now := time.Now()
	f.trackers.Range(func(key, value interface{}) bool {
		ip := key.(string)
		tracker := value.(*ipTracker)

		tracker.mu.Lock()
		// Remove general 404s older than 20s
		validGen := make([]time.Time, 0, len(tracker.general404s))
		for _, t := range tracker.general404s {
			if now.Sub(t) <= 20*time.Second {
				validGen = append(validGen, t)
			}
		}
		tracker.general404s = validGen

		// Remove scanner 404s older than 60s
		validScan := make([]time.Time, 0, len(tracker.scanner404s))
		for _, t := range tracker.scanner404s {
			if now.Sub(t) <= 60*time.Second {
				validScan = append(validScan, t)
			}
		}
		tracker.scanner404s = validScan

		empty := len(tracker.general404s) == 0 && len(tracker.scanner404s) == 0
		tracker.mu.Unlock()

		if empty {
			f.trackers.Delete(ip)
		}
		return true
	})
}

// startCleanupLoop runs a periodic 1-minute ticker to trigger tracker cleanups.
func (f *Fail2Ban) startCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			f.CleanupStaleTrackers()
		case <-f.stopChan:
			return
		}
	}
}
