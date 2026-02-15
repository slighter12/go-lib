SHELL := /bin/sh

CHANGELOG_FILE ?= CHANGELOG.md
RANGE ?= HEAD
PUSH ?= 0

.PHONY: show-release-version tag-changed-modules

show-release-version:
	@version="$(VERSION)"; \
	if [ -z "$$version" ]; then \
		version=$$(sed -nE 's/^## \[(v[0-9]+\.[0-9]+\.[0-9]+)\].*/\1/p' "$(CHANGELOG_FILE)" | head -n 1); \
	fi; \
	test -n "$$version" || (echo "Cannot resolve version from $(CHANGELOG_FILE). Set VERSION manually."; exit 1); \
	echo "$$version"

tag-changed-modules:
	@version="$(VERSION)"; \
	if [ -z "$$version" ]; then \
		version=$$(sed -nE 's/^## \[(v[0-9]+\.[0-9]+\.[0-9]+)\].*/\1/p' "$(CHANGELOG_FILE)" | head -n 1); \
	fi; \
	test -n "$$version" || (echo "Cannot resolve version from $(CHANGELOG_FILE). Set VERSION manually."; exit 1); \
	changed_files=$$(git diff-tree --no-commit-id --name-only -r "$(RANGE)"); \
	test -n "$$changed_files" || (echo "No changed files found in RANGE=$(RANGE)"; exit 1); \
	modules=$$(find database -name go.mod -exec dirname {} \; | sort); \
	tagged=0; \
	for module in $$modules; do \
		if printf '%s\n' "$$changed_files" | grep -q "^$$module/"; then \
			tag="$$module/$$version"; \
			if git rev-parse -q --verify "refs/tags/$$tag" >/dev/null; then \
				echo "Tag already exists: $$tag"; \
				exit 1; \
			fi; \
			git tag -a "$$tag" -m "Release $$tag"; \
			echo "Created tag: $$tag"; \
			if [ "$(PUSH)" = "1" ]; then \
				git push origin "$$tag"; \
				echo "Pushed tag: $$tag"; \
			fi; \
			tagged=1; \
		fi; \
	done; \
	test "$$tagged" = "1" || (echo "No changed modules with go.mod found in RANGE=$(RANGE)"; exit 1)
