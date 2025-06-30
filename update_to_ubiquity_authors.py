#!/usr/bin/env python3
"""
Script to update copyright and ownership references to "The Ubiquity Authors"
for CNCF project alignment.
"""

import os
import re
import glob
from pathlib import Path


def update_copyright_in_file(file_path):
    """Update copyright holder in a single file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # Update copyright holder from Logicalis UKI to The Ubiquity Authors
        content = re.sub(
            r'Copyright \d+ Logicalis UKI\. All Rights Reserved\.',
            'Copyright The Ubiquity Authors.',
            content
        )
        
        # Update copyright holder from just "Logicalis" to The Ubiquity Authors
        content = re.sub(
            r'Copyright \d+ Logicalis',
            'Copyright The Ubiquity Authors',
            content
        )
        
        # Only write if content changed
        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"Updated copyright in: {file_path}")
            return True
        
        return False
        
    except Exception as e:
        print(f"Error updating {file_path}: {e}")
        return False


def main():
    """Main function to update all files"""
    base_dir = Path("/Users/ccoates/Documents/git/ubiquity")
    
    # File patterns to search
    patterns = [
        '**/*.py',
        '**/*.sh',
        '**/*.yaml',
        '**/*.yml',
        '**/*.tf',
        '**/*.go',
        '**/*.md',
        '**/Makefile',
        '**/*.j2'
    ]
    
    updated_files = []
    
    for pattern in patterns:
        for file_path in glob.glob(str(base_dir / pattern), recursive=True):
            # Skip certain directories and files
            if any(skip in file_path for skip in ['.git', '__pycache__', 'node_modules']):
                continue
                
            if update_copyright_in_file(file_path):
                updated_files.append(file_path)
    
    print(f"\nUpdated {len(updated_files)} files:")
    for file_path in updated_files:
        print(f"  {file_path}")


if __name__ == "__main__":
    main()
