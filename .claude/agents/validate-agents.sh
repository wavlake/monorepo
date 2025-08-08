#!/bin/bash
# Agent Validation Script - Used by subagent-maintainer-agent

set -e

AGENTS_DIR="$(dirname "$0")"
cd "$AGENTS_DIR"

echo "🔍 Validating Subagent Definitions"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Validation counters
TOTAL=0
PASSED=0
WARNINGS=0
FAILED=0

# Validate each agent file
for agent_file in *.md; do
    if [[ "$agent_file" == "README.md" ]]; then
        continue
    fi
    
    echo -e "\n📄 Validating: $agent_file"
    TOTAL=$((TOTAL + 1))
    ISSUES=""
    WARNS=""
    
    # Check required sections (YAML frontmatter)
    if ! grep -q "^name:" "$agent_file"; then
        ISSUES="${ISSUES}❌ Missing YAML frontmatter name\n"
    fi
    
    if ! grep -q "^## Purpose" "$agent_file"; then
        ISSUES="${ISSUES}❌ Missing Purpose section\n"
    fi
    
    if ! grep -q "## Core Capabilities" "$agent_file"; then
        ISSUES="${ISSUES}❌ Missing Core Capabilities section\n"
    fi
    
    if ! grep -q "## Tools Available" "$agent_file"; then
        ISSUES="${ISSUES}❌ Missing Tools Available section\n"
    fi
    
    # Check for dangerous commands
    if grep -qE "rm -rf|sudo|chmod 777" "$agent_file"; then
        ISSUES="${ISSUES}❌ Contains potentially dangerous commands\n"
    fi
    
    # Validate task commands exist
    task_commands=$(grep -oE "task [a-z:-]+" "$agent_file" | cut -d' ' -f2 | sort -u || true)
    if [[ -n "$task_commands" ]]; then
        while IFS= read -r cmd; do
            if ! task --list 2>/dev/null | grep -q "^  $cmd:"; then
                WARNS="${WARNS}⚠️  Task command not found: $cmd\n"
            fi
        done <<< "$task_commands"
    fi
    
    # Validate file paths
    paths=$(grep -oE "apps/[^/]+/|packages/[^/]+/" "$agent_file" | sort -u || true)
    if [[ -n "$paths" ]]; then
        while IFS= read -r path; do
            if [[ ! -d "../../$path" ]]; then
                WARNS="${WARNS}⚠️  Path not found: $path\n"
            fi
        done <<< "$paths"
    fi
    
    # Check for test validation requirement
    if [[ "$agent_file" != "subagent-maintainer-agent.md" ]]; then
        if ! grep -qE "test.*validation|validate.*test|ensure.*test.*pass" "$agent_file"; then
            WARNS="${WARNS}⚠️  No explicit test validation requirement found\n"
        fi
    fi
    
    # Report results
    if [[ -z "$ISSUES" ]] && [[ -z "$WARNS" ]]; then
        echo -e "${GREEN}✅ PASSED${NC}"
        PASSED=$((PASSED + 1))
    elif [[ -z "$ISSUES" ]] && [[ -n "$WARNS" ]]; then
        echo -e "${YELLOW}⚠️  WARNINGS${NC}"
        echo -e "$WARNS"
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo -e "$ISSUES"
        [[ -n "$WARNS" ]] && echo -e "$WARNS"
        FAILED=$((FAILED + 1))
    fi
done

# Summary
echo -e "\n================================="
echo "📊 Validation Summary"
echo "================================="
echo -e "Total Agents: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

# Exit with appropriate code
if [[ $FAILED -gt 0 ]]; then
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    exit 0  # Warnings don't fail the validation
else
    exit 0
fi