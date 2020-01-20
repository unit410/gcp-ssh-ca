#! /usr/bin/env bash

echo "Try to pull signed SSH key down from the metadata service for 60s"
for i in {1..60}; do
  # Check the Metadata Service for a signed key
  signed_key=$(curl --silent -H "Metadata-Flavor: Google" \
    "http://metadata.google.internal/computeMetadata/v1/instance/attributes/hostkeys-signed-ssh-ed25519")
  if [[ $signed_key == "ssh-ed25519-cert"* ]]; then
    echo "Signed key found.  Saving to disk"
    # Save to disk
    echo $signed_key > /etc/ssh/ssh_host_ed25519_key-cert.pub
    # Configure sshd_config to use certs
    echo 'HostCertificate /etc/ssh/ssh_host_ed25519_key-cert.pub' >> /etc/ssh/sshd_config
    # Reload sshd
    systemctl reload sshd
    break
  fi
  sleep 1
done

echo "Startup Complete"