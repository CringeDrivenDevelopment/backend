#!/usr/bin/env python3
"""
Telegram Init Data Generator
Generates mock Telegram WebApp init data with valid hash signature.
"""

import hmac
import hashlib
import urllib.parse
import json
import time
import os
from typing import Dict, Any

# Mock data configuration
MOCK_USER_DATA = {
    "id": 687627953,
    "first_name": "linuxfight",
    "last_name": "",  # Empty last name
    "username": "linuxfight",
    "language_code": "ru",
    "is_premium": False,
    "allows_write_to_pm": False,
    "photo_url": "https://example.com/cat.jpeg"
}

def load_bot_token() -> str:
    """
    Load bot token from dev.env file.
    Expected format in dev.env: BOT_TOKEN=your_token_here
    """
    env_file = "dev.env"

    if not os.path.exists(env_file):
        raise FileNotFoundError(f"{env_file} file not found. Please create it with BOT_TOKEN=your_token")

    with open(env_file, 'r') as file:
        for line in file:
            line = line.strip()
            if line.startswith('BOT_TOKEN='):
                token = line.split('=', 1)[1].strip()
                if not token:
                    raise ValueError("BOT_TOKEN is empty in dev.env file")
                return token

    raise ValueError("BOT_TOKEN not found in dev.env file")

def create_user_json(user_data: Dict[str, Any]) -> str:
    """
    Create user JSON string for Telegram init data.
    Removes empty string values to match Telegram's behavior.
    """
    # Remove empty string values (but keep False boolean values)
    filtered_data = {
        k: v for k, v in user_data.items()
        if v != "" and v is not None
    }

    # Convert to JSON without spaces (compact format)
    return json.dumps(filtered_data, separators=(',', ':'))

def generate_init_data_hash(data_dict: Dict[str, str], bot_token: str) -> str:
    """
    Generate hash for Telegram init data using the official algorithm:
    1. Create array of key=value pairs (excluding hash)
    2. Sort alphabetically
    3. Create HMAC-SHA256 with key "WebAppData" and bot token
    4. Create HMAC-SHA256 with previous result and joined pairs
    5. Return as hex string
    """
    # Step 1: Create key=value pairs (excluding hash)
    pairs = []
    for key, value in data_dict.items():
        if key != 'hash':
            pairs.append(f"{key}={value}")

    # Step 2: Sort alphabetically
    pairs.sort()

    # Step 3: Create HMAC-SHA256 with "WebAppData" and bot token
    secret_key = hmac.new(
        "WebAppData".encode('utf-8'),
        bot_token.encode('utf-8'),
        hashlib.sha256
    ).digest()

    # Step 4: Create HMAC-SHA256 with secret key and joined pairs
    data_string = "\n".join(pairs)
    hash_value = hmac.new(
        secret_key,
        data_string.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()

    return hash_value

def generate_mock_init_data(bot_token: str,
                            user_data: Dict[str, Any] = None,
                            auth_date: int = None) -> str:
    """
    Generate complete mock Telegram init data string.

    Args:
        bot_token: Telegram bot token for hash generation
        user_data: Custom user data (uses MOCK_USER_DATA if None)
        auth_date: Unix timestamp for auth_date (uses current time if None)

    Returns:
        URL-encoded init data string
    """
    if user_data is None:
        user_data = MOCK_USER_DATA.copy()

    if auth_date is None:
        auth_date = int(time.time())

    # Create user JSON
    user_json = create_user_json(user_data)

    # Create init data parameters (without hash)
    init_data = {
        "user": user_json,
        "chat_type": "private",
        "auth_date": str(auth_date),
    }

    # Generate hash using the algorithm
    hash_value = generate_init_data_hash(init_data, bot_token)

    # Add hash to init data
    init_data["hash"] = hash_value

    # Convert to URL query string
    query_params = []
    for key, value in init_data.items():
        encoded_value = urllib.parse.quote(str(value), safe='')
        query_params.append(f"{key}={encoded_value}")

    return "&".join(query_params)

def parse_and_display_init_data(init_data_string: str) -> None:
    """
    Parse and display init data in a readable format.
    """
    parsed = urllib.parse.parse_qs(init_data_string)

    print("=== Parsed Init Data ===")
    for key, value_list in parsed.items():
        value = value_list[0]  # parse_qs returns lists

        if key == "user":
            try:
                # Try to parse and pretty-print user JSON
                user_data = json.loads(value)
                print(f"{key}:")
                for user_key, user_value in user_data.items():
                    print(f"  {user_key}: {user_value}")
            except json.JSONDecodeError:
                print(f"{key}: {value}")
        else:
            print(f"{key}: {value}")

def main():
    """
    Main function that generates mock init data once using bot token from dev.env
    """
    try:
        # Load bot token from dev.env file
        bot_token = load_bot_token()
        print("✅ Bot token loaded from dev.env")

        # Generate mock init data
        print("📝 Generating mock Telegram init data...")
        mock_init_data = generate_mock_init_data(bot_token)

        print("\n🎯 Generated Init Data:")
        print("=" * 80)
        print(mock_init_data)
        print("=" * 80)

        # Parse and display in readable format
        print("\n📊 Parsed Data:")
        print("-" * 50)
        parse_and_display_init_data(mock_init_data)

    except FileNotFoundError as e:
        print(f"❌ Error: {e}")
        print("\n💡 Create a dev.env file with:")
        print("BOT_TOKEN=your_telegram_bot_token_here")
    except ValueError as e:
        print(f"❌ Error: {e}")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")

if __name__ == "__main__":
    main()