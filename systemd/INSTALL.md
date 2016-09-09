# system.d install

1. Install the `bmp180reporter` binary to `/usr/local/bin`
2. Create the `/etc/bmp180reporter.toml` config file. See the [sample file](../config.sample.toml) for details.
3. Copy the `bmp180reporter` systemd file to `/etc/default`
4. Copy the `bmp180reporter.service` systemd file to `/etc/systemd/system`
5. Then enable and run the service:

  ```sh
  systemctl daemon-reload
  systemctl enable bmp180reporter
  systemctl start bmp180reporter
  ```

You can check the status with `systemctl status bmp180reporter`.

Run `journalctl -u bmp180reporter` to see the logs.
