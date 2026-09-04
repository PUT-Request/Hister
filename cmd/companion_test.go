package cmd

import "testing"

func TestQutebrowserCompanionCommandRegistration(t *testing.T) {
	companion, _, err := rootCmd.Find([]string{"companion"})
	if err != nil {
		t.Fatal(err)
	}
	if companion != companionCmd {
		t.Fatalf("companion command = %q, want registered companion command", companion.Name())
	}

	qutebrowser, _, err := rootCmd.Find([]string{"companion", "qutebrowser"})
	if err != nil {
		t.Fatal(err)
	}
	if qutebrowser != companionQutebrowserCmd {
		t.Fatalf(
			"qutebrowser command = %q, want registered qutebrowser companion",
			qutebrowser.Name(),
		)
	}

	for _, flagName := range []string{
		"command-timeout",
		"debounce",
		"devtools-url",
		"initial-delay",
		"label",
		"max-favicon-bytes",
		"max-wait",
		"reconnect-delay",
		"request-timeout",
		"retry-delay",
	} {
		if qutebrowser.Flags().Lookup(flagName) == nil {
			t.Errorf("qutebrowser companion is missing --%s", flagName)
		}
	}
}
