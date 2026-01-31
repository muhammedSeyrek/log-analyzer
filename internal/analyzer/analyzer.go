package analyzer

import (
	"bufio"
	"fmt"
	"log-analyzer/internal/config"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Result struct {
	Line      string
	RuleName  string
	RuleType  string
	Timestamp time.Time
}

func AnalyzeFile(filePath string, rules []config.Rule) ([]Result, error) {

	var scanner *bufio.Scanner
	var cmd *exec.Cmd
	var file *os.File

	if runtime.GOOS == "windows" && (filePath == "SECURITY_LOGS" || filePath == "SYSTEM") {
		psCommand := "Get-EventLog -LogName Security -Newest 500 | Select-Object -ExpandProperty Message"
		cmd = exec.Command("powershell", "-Command", psCommand)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, err
		}
		scanner = bufio.NewScanner(stdout)

	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("log", "show", "--style", "syslog", "--last", "10m")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, err
		}
		scanner = bufio.NewScanner(stdout)

	} else {
		useJournal := false

		// If Linux and SYSTEM, check for file existence, else use Journalctl
		if runtime.GOOS == "linux" && filePath == "SYSTEM" {
			detectedFile := detectLinuxLogFile()
			if detectedFile != "" {
				filePath = detectedFile
			} else {
				// No file found, switch to Journalctl
				useJournal = true
			}
		}

		if useJournal {
			fmt.Println("-> Linux (Journald) detected. Reading logs via 'journalctl'...")
			// Fetch last 500 logs
			cmd = exec.Command("journalctl", "-n", "500", "--no-pager")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return nil, err
			}
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			scanner = bufio.NewScanner(stdout)
		} else {
			// Classic file reading mode
			fmt.Printf("-> Opening file for static analysis: %s\n", filePath)
			var err error
			file, err = os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("file not found or permission denied: %s", filePath)
			}
			scanner = bufio.NewScanner(file)
		}
	}

	var result []Result

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		for i := 0; i < len(rules); i++ {
			if strings.Contains(strings.ToLower(line), strings.ToLower(rules[i].Pattern)) {
				found := Result{
					Line:      line,
					RuleName:  rules[i].Name,
					RuleType:  rules[i].Type,
					Timestamp: time.Now(),
				}
				result = append(result, found)
			}
		}
	}

	if cmd != nil {
		cmd.Wait()
	}
	if file != nil {
		file.Close()
	}
	return result, scanner.Err()
}

// Both for Windows and Linux
func WatchFile(filePath string, rules []config.Rule) error {

	if runtime.GOOS == "windows" && (filePath == "SECURITY_LOGS" || filePath == "SYSTEM") {
		return watchWindowsLogs(rules)
	} else if runtime.GOOS == "darwin" {
		return watchMacLogs(rules)
	} else {
		if runtime.GOOS == "linux" && filePath == "SYSTEM" {
			detectedFile := detectLinuxLogFile()
			if detectedFile != "" {
				return watchLinuxLogs(detectedFile, rules)
			} else {
				// If no file found, listen to Journalctl
				return watchLinuxJournal(rules)
			}
		}
		return watchLinuxLogs(filePath, rules)
	}
}

func watchWindowsLogs(rules []config.Rule) error {
	fmt.Println("Starting live monitoring of Windows Security Logs (Robust Mode)...")

	psCommand := `
		$last = Get-Date
		while ($true) {
			Start-Sleep -Seconds 1
			try {
				$events = Get-EventLog -LogName Security -After $last -ErrorAction Stop
				$events | ForEach-Object { 
					$_.Message
					$last = Get-Date
				}
			} catch {
				# Yeni log yoksa veya hata oluşursa sessizce devam et
			}
		}
	`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("[PS ERROR]:", scanner.Text())
		}
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		checkRules(scanner.Text(), rules)
	}

	return cmd.Wait()
}

func watchLinuxLogs(filePath string, rules []config.Rule) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(0, 2) // Move to the end of the file
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	fmt.Println("Starting live monitoring of file:", filePath)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		checkRules(line, rules)
	}
}

func watchMacLogs(rules []config.Rule) error {
	// "log stream" provides live updates
	cmd := exec.Command("log", "stream", "--style", "syslog", "--level", "info")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		checkRules(scanner.Text(), rules)
	}
	return cmd.Wait()
}

func checkRules(line string, rules []config.Rule) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	for i := 0; i < len(rules); i++ {
		if strings.Contains(strings.ToLower(line), strings.ToLower(rules[i].Pattern)) {
			fmt.Printf("\n Alert %s Detected!\n", rules[i].Name)
			fmt.Printf(" -> %s\n", line)
		}
	}
}

func detectLinuxLogFile() string {
	candidates := []string{
		"/var/log/syslog",   // Debian/Ubuntu
		"/var/log/auth.log", // Debian/Kali (Security)
		"/var/log/messages", // RHEL/CentOS
		"/var/log/secure",   // RHEL (Security)
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "" // Return empty if none found
}

func watchLinuxJournal(rules []config.Rule) error {
	fmt.Println("Starting live monitoring via 'journalctl -f'...")

	cmd := exec.Command("journalctl", "-f", "-n", "0")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		checkRules(scanner.Text(), rules)
	}
	return cmd.Wait()
}
