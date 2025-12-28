class CksWeightRoom < Formula
  desc "Kubernetes Security Practice Environment"
  homepage "https://github.com/patrickvassell/cks-weight-room"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/patrickvassell/cks-weight-room/releases/download/v#{version}/cks-weight-room-darwin-arm64"
      sha256 "SHA256_PLACEHOLDER_ARM64" # Update with actual SHA256
    else
      url "https://github.com/patrickvassell/cks-weight-room/releases/download/v#{version}/cks-weight-room-darwin-amd64"
      sha256 "SHA256_PLACEHOLDER_AMD64" # Update with actual SHA256
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/patrickvassell/cks-weight-room/releases/download/v#{version}/cks-weight-room-linux-arm64"
      sha256 "SHA256_PLACEHOLDER_LINUX_ARM64" # Update with actual SHA256
    else
      url "https://github.com/patrickvassell/cks-weight-room/releases/download/v#{version}/cks-weight-room-linux-amd64"
      sha256 "SHA256_PLACEHOLDER_LINUX_AMD64" # Update with actual SHA256
    end
  end

  depends_on "docker"

  def install
    # Download URL pattern: cks-weight-room-{os}-{arch}
    binary_name = if OS.mac?
      Hardware::CPU.arm? ? "cks-weight-room-darwin-arm64" : "cks-weight-room-darwin-amd64"
    else
      Hardware::CPU.arm? ? "cks-weight-room-linux-arm64" : "cks-weight-room-linux-amd64"
    end

    # Rename to cks-weight-room and install
    bin.install binary_name => "cks-weight-room"
  end

  test do
    assert_match(/CKS Weight Room v/, shell_output("#{bin}/cks-weight-room --version"))
  end
end
