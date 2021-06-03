# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class HttpsignatureProxy < Formula
  desc "Localhost HTTP Signatures proxy."
  homepage "https://github.com/upvestco/httpsignature-proxy"
  version "1.1.1"
  license "Apache 2.0"
  bottle :unneeded

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/upvestco/httpsignature-proxy/releases/download/v1.1.1/httpsignature-proxy_v1.1.1_macOS_64-bit.tar.gz"
      sha256 "32a72ee8043c9be7a2adbfec4d32f452d1ec1c33c22decdc895b1a90e4182536"
    end
    if Hardware::CPU.arm?
      url "https://github.com/upvestco/httpsignature-proxy/releases/download/v1.1.1/httpsignature-proxy_v1.1.1_macOS_arm64.tar.gz"
      sha256 "57949df7ce307e023cb1090ea5fc6d4aeaa1255990b246f11aeadd323e52edee"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/upvestco/httpsignature-proxy/releases/download/v1.1.1/httpsignature-proxy_v1.1.1_Linux_64-bit.tar.gz"
      sha256 "7c79170693764d9acd4462c3311fae2794d001d34207bf158d3cdf6dfc152aa6"
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/upvestco/httpsignature-proxy/releases/download/v1.1.1/httpsignature-proxy_v1.1.1_Linux_arm64.tar.gz"
      sha256 "7593d08258ab9bb4b462708dfbef260f09edf546f9fafd7ef7db0424f5895470"
    end
  end

  def install
    bin.install "httpsignature-proxy"
  end

  test do
    system "#{bin}/httpsignature-proxy version"
  end
end
