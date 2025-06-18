import os
import shutil
from pathlib import Path
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
import base64
from hashlib import pbkdf2_hmac
import sys

# 配置项
SRC_EXT = ".go"
ENC_EXT = ".go.enc"
IGNORED_FILES = ['.git', '.gitignore', 'encrypt_push.py']
FIXED_SALT = b"Fucking_Gov"

def get_key_from_password(password: str, salt: bytes = FIXED_SALT, iterations: int = 100000) -> bytes:
    return pbkdf2_hmac('sha256', password.encode('utf-8'), salt, iterations, dklen=32)

def aes_encrypt(content, key):
    cipher = AES.new(key, AES.MODE_CBC, iv=key[:16])
    ct_bytes = cipher.encrypt(pad(content.encode('utf-8'), AES.block_size))
    return base64.b64encode(cipher.iv + ct_bytes)

def get_go_files(directory):
    return [f for f in Path(directory).rglob(f'*{SRC_EXT}') if f.is_file()]

def clean_unencrypted_files():
    for item in os.listdir('.'):
        if item in IGNORED_FILES:
            continue
        path = Path(item)
        if path.is_dir():
            shutil.rmtree(path)
        elif not path.name.endswith(ENC_EXT):
            path.unlink()

def encrypt_go_files(files, key):
    encrypted = []
    for file in files:
        with open(file, 'r', encoding='utf-8') as f:
            raw = f.read()
        enc = aes_encrypt(raw, key)
        enc_path = f"{file}{ENC_EXT}"
        encrypted.append((enc_path, enc))
    return encrypted

def write_encrypted_files(encrypted_files):
    for enc_path, enc_content in encrypted_files:
        os.makedirs(os.path.dirname(enc_path), exist_ok=True)
        with open(enc_path, 'wb') as f:
            f.write(enc_content)

def create_gitignore():
    with open(".gitignore", "w") as f:
        f.write("*.go\n!*.go.enc\n")

def main():
    if len(sys.argv) >= 2:
        password = sys.argv[1]
    else:
        password = "Sm1leSi0n"

    key = get_key_from_password(password)

    print("🔍 查找 .go 文件...")
    go_files = get_go_files(".")
    print(f"🔐 加密 {len(go_files)} 个 .go 文件...")
    encrypted_files = encrypt_go_files(go_files, key)

    print("🧹 清除未加密文件（保留 .go.enc）...")
    clean_unencrypted_files()

    print("💾 写入加密文件...")
    write_encrypted_files(encrypted_files)

    print("📄 写入 .gitignore 忽略原始 .go 文件...")
    create_gitignore()

    print("✅ 加密完成，请手动 git add/commit/push。")

if __name__ == "__main__":
    main()
