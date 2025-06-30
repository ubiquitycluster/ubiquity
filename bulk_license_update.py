#!/usr/bin/env python3

import os
import re
import glob

def update_license_in_file(file_path):
    """Update license header in a single file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Skip if file doesn't contain copyright header
        if "Copyright" not in content:
            return False
        
        print(f"Updating {file_path}")
        
        # Update copyright holder from old organization to The Ubiquity Authors
        content = re.sub(
            r'Copyright \d+ Logicalis UKI\. All Rights Reserved\.',
            'Copyright The Ubiquity Authors.',
            content
        )
        
        # Update the license line
        content = re.sub(
            r'Licensed under the Functional Source License, Version 1\.0, Apache 2\.0 Change License',
            'Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License',
            content
        )
        
        # Update the URL
        content = re.sub(
            r'https://github\.com/ubiquitycluster/ubiquity/blob/main/LICENSE\.md',
            'https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE',
            content
        )
        
        # Update the transition language
        content = re.sub(
            r'It also allows for the transition of this software to an Apache 2\.0 Licence',
            'This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License',
            content
        )
        
        # Update the date reference
        content = re.sub(
            r'on the second anniversary of the date we make the software available\.',
            'as of June 2025.',
            content
        )
        
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        return True
        
    except Exception as e:
        print(f"Error updating {file_path}: {e}")
        return False

def main():
    # Find all files that might contain license headers
    patterns = [
        '**/*.py',
        '**/*.go', 
        '**/*.tf',
        '**/*.yaml',
        '**/*.yml',
        '**/*.sh',
        '**/Makefile'
    ]
    
    files_updated = 0
    
    for pattern in patterns:
        for file_path in glob.glob(pattern, recursive=True):
            if update_license_in_file(file_path):
                files_updated += 1
    
    print(f"\nTotal files updated: {files_updated}")

if __name__ == "__main__":
    main()
