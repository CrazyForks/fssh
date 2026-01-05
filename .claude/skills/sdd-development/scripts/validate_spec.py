#!/usr/bin/env python3
"""
Specification Validation Script

Validates SDD specification and plan documents for completeness,
constitutional compliance, and quality standards.

Usage:
    python validate_spec.py <path-to-spec-directory>
    
Example:
    python validate_spec.py specs/003-chat-system/
"""

import sys
import os
import re
from pathlib import Path
from typing import List, Tuple, Dict

class ValidationResult:
    def __init__(self):
        self.errors: List[str] = []
        self.warnings: List[str] = []
        self.info: List[str] = []
    
    def add_error(self, message: str):
        self.errors.append(f"❌ ERROR: {message}")
    
    def add_warning(self, message: str):
        self.warnings.append(f"⚠️  WARNING: {message}")
    
    def add_info(self, message: str):
        self.info.append(f"ℹ️  INFO: {message}")
    
    def is_valid(self) -> bool:
        return len(self.errors) == 0
    
    def print_report(self):
        print("\n" + "="*70)
        print("SPECIFICATION VALIDATION REPORT")
        print("="*70 + "\n")
        
        if self.errors:
            print("ERRORS:")
            for error in self.errors:
                print(f"  {error}")
            print()
        
        if self.warnings:
            print("WARNINGS:")
            for warning in self.warnings:
                print(f"  {warning}")
            print()
        
        if self.info:
            print("INFO:")
            for info in self.info:
                print(f"  {info}")
            print()
        
        print("="*70)
        if self.is_valid():
            print("✅ VALIDATION PASSED")
        else:
            print("❌ VALIDATION FAILED")
        print("="*70 + "\n")

def read_file(path: Path) -> str:
    """Read file content, return empty string if file doesn't exist."""
    try:
        return path.read_text()
    except FileNotFoundError:
        return ""

def count_needs_clarification(content: str) -> int:
    """Count [NEEDS CLARIFICATION] tags in content."""
    return len(re.findall(r'\[NEEDS CLARIFICATION[:\]]', content, re.IGNORECASE))

def validate_spec_md(spec_path: Path, result: ValidationResult):
    """Validate spec.md file."""
    content = read_file(spec_path)
    
    if not content:
        result.add_error("spec.md file not found")
        return
    
    # Check for unresolved clarifications
    clarification_count = count_needs_clarification(content)
    if clarification_count > 0:
        result.add_error(f"spec.md contains {clarification_count} unresolved [NEEDS CLARIFICATION] tags")
    
    # Check for required sections
    required_sections = [
        "Overview",
        "User Stories",
        "Functional Requirements",
        "Non-Functional Requirements",
        "Success Criteria"
    ]
    
    for section in required_sections:
        if f"## {section}" not in content and f"# {section}" not in content:
            result.add_warning(f"spec.md missing recommended section: {section}")
    
    # Check for acceptance criteria
    if "Acceptance Criteria" not in content:
        result.add_warning("spec.md should include Acceptance Criteria for user stories")
    
    # Check that it focuses on WHAT/WHY, not HOW
    technical_terms = [
        r'\bAPI\b', r'\bREST\b', r'\bWebSocket\b', r'\bPostgreSQL\b', 
        r'\bRedis\b', r'\bMongoDB\b', r'\breact\b', r'\bvue\b', r'\bangular\b'
    ]
    
    for term in technical_terms:
        if re.search(term, content, re.IGNORECASE):
            result.add_warning(
                f"spec.md contains technical implementation term ('{term}'). "
                "Specifications should focus on WHAT/WHY, not HOW."
            )
            break
    
    result.add_info(f"spec.md found and validated ({len(content)} characters)")

def validate_plan_md(plan_path: Path, result: ValidationResult):
    """Validate plan.md file."""
    content = read_file(plan_path)
    
    if not content:
        result.add_warning("plan.md file not found (may not be created yet)")
        return
    
    # Check for constitutional gates
    required_gates = [
        "Simplicity Gate",
        "Anti-Abstraction Gate",
        "Integration-First Gate",
        "Test-First Gate"
    ]
    
    for gate in required_gates:
        if gate not in content:
            result.add_error(f"plan.md missing required constitutional gate: {gate}")
    
    # Check for ≤3 projects rule
    if "Number of projects:" in content:
        match = re.search(r'Number of projects:\s*\[?(\d+)\]?', content)
        if match:
            num_projects = int(match.group(1))
            if num_projects > 3:
                if "Complexity Tracking" not in content:
                    result.add_error(
                        f"plan.md has {num_projects} projects (>3) but no "
                        "Complexity Tracking section justifying this"
                    )
                else:
                    result.add_warning(
                        f"plan.md uses {num_projects} projects (>3). "
                        "Verify justification in Complexity Tracking section."
                    )
    
    # Check for test-first approach
    if "Test Creation Order" not in content and "Testing Strategy" not in content:
        result.add_error("plan.md missing Testing Strategy or Test Creation Order")
    
    # Check for traceability
    if "Traceability" not in content:
        result.add_warning(
            "plan.md should include traceability matrix mapping requirements to implementation"
        )
    
    result.add_info(f"plan.md found and validated ({len(content)} characters)")

def validate_supporting_files(spec_dir: Path, result: ValidationResult):
    """Validate supporting documentation files."""
    
    # Check for data-model.md
    data_model = spec_dir / "data-model.md"
    if data_model.exists():
        result.add_info("data-model.md found")
    else:
        result.add_warning("data-model.md not found (may not be needed for this feature)")
    
    # Check for contracts directory
    contracts_dir = spec_dir / "contracts"
    if contracts_dir.exists() and contracts_dir.is_dir():
        contract_files = list(contracts_dir.glob("*"))
        if contract_files:
            result.add_info(f"contracts/ directory found with {len(contract_files)} files")
        else:
            result.add_warning("contracts/ directory exists but is empty")
    else:
        result.add_warning("contracts/ directory not found (may not be needed for this feature)")
    
    # Check for research.md
    research = spec_dir / "research.md"
    if research.exists():
        result.add_info("research.md found")
    
    # Check for quickstart.md
    quickstart = spec_dir / "quickstart.md"
    if quickstart.exists():
        result.add_info("quickstart.md found")
    else:
        result.add_warning("quickstart.md not found (should contain key validation scenarios)")
    
    # Check for tasks.md
    tasks = spec_dir / "tasks.md"
    if tasks.exists():
        result.add_info("tasks.md found")

def validate_specification(spec_dir_path: str) -> ValidationResult:
    """Main validation function."""
    result = ValidationResult()
    spec_dir = Path(spec_dir_path)
    
    if not spec_dir.exists():
        result.add_error(f"Directory not found: {spec_dir_path}")
        return result
    
    if not spec_dir.is_dir():
        result.add_error(f"Path is not a directory: {spec_dir_path}")
        return result
    
    # Validate spec.md
    spec_md = spec_dir / "spec.md"
    validate_spec_md(spec_md, result)
    
    # Validate plan.md
    plan_md = spec_dir / "plan.md"
    validate_plan_md(plan_md, result)
    
    # Validate supporting files
    validate_supporting_files(spec_dir, result)
    
    return result

def main():
    if len(sys.argv) < 2:
        print("Usage: python validate_spec.py <path-to-spec-directory>")
        print("Example: python validate_spec.py specs/003-chat-system/")
        sys.exit(1)
    
    spec_dir = sys.argv[1]
    result = validate_specification(spec_dir)
    result.print_report()
    
    # Exit with error code if validation failed
    sys.exit(0 if result.is_valid() else 1)

if __name__ == "__main__":
    main()
