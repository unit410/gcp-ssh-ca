provider "google" {
  project     = "guest-attributes-test"
  region      = "us-west1"
}

resource "google_compute_instance" "instance" {
  name = "instance"
  zone = "us-west1-a"

  machine_type = "g1-small"

  network_interface {
    network = "default"

    access_config {
    }
  }

  boot_disk {
    initialize_params {
      size  = 10
      type  = "pd-standard"
      image = "gce-uefi-images/ubuntu-1804-lts"    
    }
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  metadata = {
    startup-script = file("./startup.sh")
    enable-guest-attributes = "TRUE"
  }

  # Metadata will be updated by our CA signer.
  lifecycle { ignore_changes = [metadata] }
}