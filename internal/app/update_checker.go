package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/laborin/pico-xbar/internal/statusbar"
	"github.com/laborin/pico-xbar/internal/version"
)

const (
	updateCheckInterval = 12 * time.Hour
	updateCheckTimeout  = 15 * time.Second
	releasesAPIURL      = "https://api.github.com/repos/laborin/pico-xbar/releases/latest"
	defaultReleasesURL  = "https://github.com/laborin/pico-xbar/releases"
)

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func (a *App) startUpdateCheckLoop() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.runUpdateCheckLoop(a.ctx)
	}()
}

func (a *App) runUpdateCheckLoop(ctx context.Context) {
	currentVersion := normalizeVersion(version.Version)
	if currentVersion == "" {
		return
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			updateAvailable, latestVersion, releaseURL, err := checkForUpdate(ctx, currentVersion)
			if err != nil {
				log.Printf("update check failed: %v", err)
				timer.Reset(updateCheckInterval)
				continue
			}
			if updateAvailable {
				if ctx.Err() != nil {
					return
				}
				a.showUpdateAvailableAlert(currentVersion, latestVersion, releaseURL)
				return
			}
			timer.Reset(updateCheckInterval)
		}
	}
}

func (a *App) showUpdateAvailableAlert(currentVersion, latestVersion, releaseURL string) {
	message := statusbar.LocalizedString("update_available_message")
	message = strings.Replace(message, "%@", currentVersion, 1)
	message = strings.Replace(message, "%@", latestVersion, 1)
	statusbar.ShowAlertWithURLButtons(
		statusbar.LocalizedString("Update Available"),
		message,
		statusbar.LocalizedString("Download"),
		statusbar.LocalizedString("Dismiss"),
		releaseURL,
	)
}

func checkForUpdate(ctx context.Context, currentVersion string) (bool, string, string, error) {
	requestCtx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, releasesAPIURL, nil)
	if err != nil {
		return false, "", "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "pico-xbar")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, "", "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("github releases api status: %s", response.Status)
	}

	var latestRelease latestReleaseResponse
	if err := json.NewDecoder(response.Body).Decode(&latestRelease); err != nil {
		return false, "", "", err
	}

	latestVersion := normalizeVersion(latestRelease.TagName)
	if latestVersion == "" {
		return false, "", "", fmt.Errorf("invalid latest release tag: %q", latestRelease.TagName)
	}

	releaseURL := strings.TrimSpace(latestRelease.HTMLURL)
	if releaseURL == "" {
		releaseURL = defaultReleasesURL
	}

	return isVersionNewer(latestVersion, currentVersion), latestVersion, releaseURL, nil
}

func normalizeVersion(raw string) string {
	version := strings.TrimSpace(raw)
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		return ""
	}

	cutIndex := len(version)
	for i := 0; i < len(version); i++ {
		c := version[i]
		if (c < '0' || c > '9') && c != '.' {
			cutIndex = i
			break
		}
	}
	version = strings.Trim(version[:cutIndex], ".")
	if version == "" {
		return ""
	}
	return version
}

func isVersionNewer(latest, current string) bool {
	latestParts, err := parseVersionParts(latest)
	if err != nil {
		return false
	}
	currentParts, err := parseVersionParts(current)
	if err != nil {
		return false
	}

	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}
	for i := 0; i < maxLen; i++ {
		latestPart := 0
		if i < len(latestParts) {
			latestPart = latestParts[i]
		}
		currentPart := 0
		if i < len(currentParts) {
			currentPart = currentParts[i]
		}
		if latestPart > currentPart {
			return true
		}
		if latestPart < currentPart {
			return false
		}
	}
	return false
}

func parseVersionParts(version string) ([]int, error) {
	segments := strings.Split(version, ".")
	if len(segments) == 0 {
		return nil, fmt.Errorf("empty version")
	}

	parts := make([]int, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			return nil, fmt.Errorf("invalid version: %s", version)
		}
		part, err := strconv.Atoi(segment)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}
