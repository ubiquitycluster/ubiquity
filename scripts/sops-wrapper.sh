#!/bin/bash
# scripts/sops-wrapper.sh - Password protected SOPS wrapper

set -euo pipefail

# Age key location
AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"
AGE_KEY_ENCRYPTED="$HOME/.config/sops/age/keys.txt.gpg"

# Function to unlock age key
unlock_age_key() {
    if [[ -f "$AGE_KEY_ENCRYPTED" && ! -f "$AGE_KEY_FILE" ]]; then
        echo "🔐 Unlocking SOPS age key..."
        gpg --decrypt "$AGE_KEY_ENCRYPTED" > "$AGE_KEY_FILE"
        chmod 600 "$AGE_KEY_FILE"
        
        # Set up cleanup trap
        trap 'rm -f "$AGE_KEY_FILE"' EXIT
    fi
}

# Function to lock age key  
lock_age_key() {
    if [[ -f "$AGE_KEY_FILE" && ! -f "$AGE_KEY_ENCRYPTED" ]]; then
        echo "🔒 Encrypting age key with GPG..."
        gpg --symmetric --cipher-algo AES256 --output "$AGE_KEY_ENCRYPTED" "$AGE_KEY_FILE"
        rm -f "$AGE_KEY_FILE"
    fi
}

# Main execution
case "${1:-}" in
    "unlock")
        unlock_age_key
        ;;
    "lock") 
        lock_age_key
        ;;
    *)
        unlock_age_key
        ./scripts/secure-configure.py "$@"
        ;;
esac