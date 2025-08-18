
DIFF_OUTPUT ?= tmp/diff_output.txt

# Run all tests
.PHONY: test
test:
	@echo "Testing..."
	@go test ./... || { echo "Tests failed!"; exit 1; }
	@echo "Done!"

# Show test coverage
.PHONY: coverage
coverage:
	@echo "Generating test coverage..."
	@go test -cover ./... || { echo "Coverage generation failed!"; exit 1; }
	@echo "Coverage displayed!"

# Open coverage report in browser
.PHONY: cover
cover:
	@echo "Generating and opening coverage report..."
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out || { echo "Coverage report failed!"; exit 1; }
	@echo "Coverage report opened!"

# Stage all files
.PHONY: stage-all
stage-all:
	@echo "Staging all files..."
	@git add . || { echo "Failed to stage files!"; exit 1; }
	@echo "All files staged!"

# Unstage all files
.PHONY: unstage-all
unstage-all:
	@echo "Unstaging all files..."
	@git restore --staged . || { echo "Failed to unstage files!"; exit 1; }
	@echo "All files unstaged!"

# Copy diff to clipboard
.PHONY: diff-to-clipboard
diff-to-clipboard:
	@echo "Copying diff to clipboard..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		command -v pbcopy >/dev/null 2>&1 || { echo "Error: pbcopy not found (macOS)"; exit 1; }; \
		git diff --staged | pbcopy || { echo "Failed to copy diff!"; exit 1; }; \
		echo "Diff copied (macOS)"; \
	elif [ "$$(uname)" = "Linux" ]; then \
		if command -v xclip >/dev/null 2>&1; then \
			git diff --staged | xclip -selection clipboard || { echo "Failed to copy diff!"; exit 1; }; \
			echo "Diff copied (Linux/xclip)"; \
		elif command -v wl-copy >/dev/null 2>&1; then \
			git diff --staged | wl-copy || { echo "Failed to copy diff!"; exit 1; }; \
			echo "Diff copied (Linux/wl-copy)"; \
		else \
			echo "Error: Install xclip or wl-copy for Linux clipboard support"; \
			exit 1; \
		fi; \
	elif [ "$$(uname -o 2>/dev/null)" = "Msys" ] || [ "$$(uname -o 2>/dev/null)" = "Cygwin" ]; then \
		git diff --staged | clip || { echo "Failed to copy diff!"; exit 1; }; \
		echo "Diff copied (Windows)"; \
	else \
		echo "Error: Unsupported OS for clipboard copy"; \
		exit 1; \
	fi

# Stage all, copy diff to clipboard, unstage
.PHONY: diff
diff: stage-all diff-to-clipboard unstage-all
	@echo "Diff content on clipboard and ready to paste."

# Stage all, save diff to file, unstage
.PHONY: diff-file
diff-file: stage-all
	@echo "Saving diff to ${DIFF_OUTPUT}..."
	@git diff --staged > ${DIFF_OUTPUT} || { echo "Failed to write to ${DIFF_OUTPUT}!"; exit 1; }
	@echo "Diff saved to ${DIFF_OUTPUT}."
	@$(MAKE) unstage-all

