<!-- DO NOT EDIT. This document is automatically generated. -->
# Detail Queries Summary

Following is a summary of the detail queries hardcoded in Fleet used to populate the device details:

## battery

- Platforms: darwin

- Query:

```sql
SELECT serial_number, cycle_count, health FROM battery;
```

## disk_encryption_darwin

- Platforms: darwin

- Query:

```sql
SELECT 1 FROM disk_encryption WHERE user_uuid IS NOT "" AND filevault_status = 'on' LIMIT 1;
```

## disk_encryption_linux

- Platforms: linux, ubuntu, debian, rhel, centos, sles, kali, gentoo, amzn, pop, arch, linuxmint, void, nixos

- Query:

```sql
SELECT de.encrypted, m.path FROM disk_encryption de JOIN mounts m ON m.device_alias = de.name;
```

## disk_encryption_windows

- Platforms: windows

- Query:

```sql
SELECT 1 FROM bitlocker_info WHERE drive_letter = 'C:' AND protection_status = 1;
```

## disk_space_unix

- Platforms: linux, ubuntu, debian, rhel, centos, sles, kali, gentoo, amzn, pop, arch, linuxmint, void, nixos, darwin

- Query:

```sql
SELECT (blocks_available * 100 / blocks) AS percent_disk_space_available,
       round((blocks_available * blocks_size *10e-10),2) AS gigs_disk_space_available
FROM mounts WHERE path = '/' LIMIT 1;
```

## disk_space_windows

- Platforms: windows

- Query:

```sql
SELECT ROUND((sum(free_space) * 100 * 10e-10) / (sum(size) * 10e-10)) AS percent_disk_space_available,
       ROUND(sum(free_space) * 10e-10) AS gigs_disk_space_available
FROM logical_drives WHERE file_system = 'NTFS' LIMIT 1;
```

## google_chrome_profiles

- Platforms: all

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'google_chrome_profiles';
```

- Query:

```sql
SELECT email FROM google_chrome_profiles WHERE NOT ephemeral AND email <> ''
```

## kubequery_info

- Platforms: all

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'kubernetes_info';
```

- Query:

```sql
SELECT * from kubernetes_info
```

## mdm

- Platforms: darwin

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'mdm';
```

- Query:

```sql
select enrolled, server_url, installed_from_dep, payload_identifier from mdm;
```

## mdm_windows

- Platforms: windows

- Query:

```sql
SELECT * FROM (
				SELECT "provider_id" AS "key", data as "value" FROM registry
				WHERE path LIKE 'HKEY_LOCAL_MACHINE\Software\Microsoft\Enrollments\%\ProviderID'
				LIMIT 1
			)
			UNION ALL
			SELECT * FROM (
				SELECT "discovery_service_url" AS "key", data as "value" FROM registry
				WHERE path LIKE 'HKEY_LOCAL_MACHINE\Software\Microsoft\Enrollments\%\DiscoveryServiceFullURL'
				LIMIT 1
			)
			UNION ALL
			SELECT * FROM (
				SELECT "autopilot" AS "key", 1=1 AS "value" FROM registry
				WHERE path = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Provisioning\AutopilotPolicyCache'
				LIMIT 1
			)
			UNION ALL
			SELECT * FROM (
				SELECT "installation_type" AS "key", data as "value" FROM registry
				WHERE path = 'HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\InstallationType'
				LIMIT 1
			)
			;
```

## munki_info

- Platforms: darwin

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'munki_info';
```

- Query:

```sql
select version, errors, warnings from munki_info;
```

## network_interface_chrome

- Platforms: chrome

- Query:

```sql
SELECT address, mac FROM network_interfaces LIMIT 1
```

## network_interface_unix

- Platforms: linux, ubuntu, debian, rhel, centos, sles, kali, gentoo, amzn, pop, arch, linuxmint, void, nixos, darwin

- Query:

```sql
SELECT
    ia.address,
    id.mac
FROM
    interface_addresses ia
    JOIN interface_details id ON id.interface = ia.interface
	-- On Unix ia.interface is the name of the interface,
	-- whereas on Windows ia.interface is the IP of the interface.
    JOIN routes r ON r.interface = ia.interface
WHERE
	-- Destination 0.0.0.0/0 is the default route on route tables.
    r.destination = '0.0.0.0' AND r.netmask = 0
	-- Type of route is "gateway" for Unix, "remote" for Windows.
    AND r.type = 'gateway'
	-- We are only interested on private IPs (some devices have their Public IP as Primary IP too).
    AND (
		-- Private IPv4 addresses.
		inet_aton(ia.address) IS NOT NULL AND (
			split(ia.address, '.', 0) = '10'
			OR (split(ia.address, '.', 0) = '172' AND (CAST(split(ia.address, '.', 1) AS INTEGER) & 0xf0) = 16)
			OR (split(ia.address, '.', 0) = '192' AND split(ia.address, '.', 1) = '168')
		)
		-- Private IPv6 addresses start with 'fc' or 'fd'.
		OR (inet_aton(ia.address) IS NULL AND regex_match(lower(ia.address), '^f[cd][0-9a-f][0-9a-f]:[0-9a-f:]+', 0) IS NOT NULL)
	)
ORDER BY
    r.metric ASC,
	-- Prefer IPv4 addresses over IPv6 addresses if their route have the same metric.
	inet_aton(ia.address) IS NOT NULL DESC
LIMIT 1;
```

## network_interface_windows

- Platforms: windows

- Query:

```sql
SELECT
    ia.address,
    id.mac
FROM
    interface_addresses ia
    JOIN interface_details id ON id.interface = ia.interface
	-- On Unix ia.interface is the name of the interface,
	-- whereas on Windows ia.interface is the IP of the interface.
    JOIN routes r ON r.interface = ia.address
WHERE
	-- Destination 0.0.0.0/0 is the default route on route tables.
    r.destination = '0.0.0.0' AND r.netmask = 0
	-- Type of route is "gateway" for Unix, "remote" for Windows.
    AND r.type = 'remote'
	-- We are only interested on private IPs (some devices have their Public IP as Primary IP too).
    AND (
		-- Private IPv4 addresses.
		inet_aton(ia.address) IS NOT NULL AND (
			split(ia.address, '.', 0) = '10'
			OR (split(ia.address, '.', 0) = '172' AND (CAST(split(ia.address, '.', 1) AS INTEGER) & 0xf0) = 16)
			OR (split(ia.address, '.', 0) = '192' AND split(ia.address, '.', 1) = '168')
		)
		-- Private IPv6 addresses start with 'fc' or 'fd'.
		OR (inet_aton(ia.address) IS NULL AND regex_match(lower(ia.address), '^f[cd][0-9a-f][0-9a-f]:[0-9a-f:]+', 0) IS NOT NULL)
	)
ORDER BY
    r.metric ASC,
	-- Prefer IPv4 addresses over IPv6 addresses if their route have the same metric.
	inet_aton(ia.address) IS NOT NULL DESC
LIMIT 1;
```

## orbit_info

- Platforms: all

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'orbit_info';
```

- Query:

```sql
SELECT version FROM orbit_info
```

## os_unix_like

- Platforms: linux, ubuntu, debian, rhel, centos, sles, kali, gentoo, amzn, pop, arch, linuxmint, void, nixos, darwin

- Query:

```sql
SELECT
		os.name,
		os.major,
		os.minor,
		os.patch,
		os.build,
		os.arch,
		os.platform,
		os.version AS version,
		k.version AS kernel_version
	FROM
		os_version os,
		kernel_info k
```

## os_version

- Platforms: all

- Query:

```sql
SELECT * FROM os_version LIMIT 1
```

## os_version_windows

- Platforms: windows

- Query:

```sql
SELECT
		os.name,
		os.codename as display_version
	
	FROM
		os_version os
```

## os_windows

- Platforms: windows

- Query:

```sql
SELECT
		os.name,
		os.platform,
		os.arch,
		k.version as kernel_version,
		os.codename as display_version
	
	FROM
		os_version os,
		kernel_info k
```

## osquery_flags

- Platforms: all

- Query:

```sql
select name, value from osquery_flags where name in ("distributed_interval", "config_tls_refresh", "config_refresh", "logger_tls_period")
```

## osquery_info

- Platforms: all

- Query:

```sql
select * from osquery_info limit 1
```

## scheduled_query_stats

- Platforms: all

- Query:

```sql
SELECT *,
				(SELECT value from osquery_flags where name = 'pack_delimiter') AS delimiter
			FROM osquery_schedule
```

## software_linux

- Platforms: linux, ubuntu, debian, rhel, centos, sles, kali, gentoo, amzn, pop, arch, linuxmint, void, nixos

- Query:

```sql
WITH cached_users AS (WITH cached_groups AS (select * from groups)
 SELECT uid, username, type, groupname, shell
 FROM users LEFT JOIN cached_groups USING (gid)
 WHERE type <> 'special' AND shell NOT LIKE '%/false' AND shell NOT LIKE '%/nologin' AND shell NOT LIKE '%/shutdown' AND shell NOT LIKE '%/halt' AND username NOT LIKE '%$' AND username NOT LIKE '\_%' ESCAPE '\' AND NOT (username = 'sync' AND shell ='/bin/sync' AND directory <> ''))
SELECT
  name AS name,
  version AS version,
  'Package (deb)' AS type,
  'deb_packages' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM deb_packages
WHERE status = 'install ok installed'
UNION
SELECT
  package AS name,
  version AS version,
  'Package (Portage)' AS type,
  'portage_packages' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM portage_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Package (RPM)' AS type,
  'rpm_packages' AS source,
  release AS release,
  vendor AS vendor,
  arch AS arch
FROM rpm_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Package (NPM)' AS type,
  'npm_packages' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM npm_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Chrome)' AS type,
  'chrome_extensions' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM cached_users CROSS JOIN chrome_extensions USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Firefox)' AS type,
  'firefox_addons' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM cached_users CROSS JOIN firefox_addons USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Atom)' AS type,
  'atom_packages' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM cached_users CROSS JOIN atom_packages USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Python)' AS type,
  'python_packages' AS source,
  '' AS release,
  '' AS vendor,
  '' AS arch
FROM python_packages;
```

## software_macos

- Platforms: darwin

- Query:

```sql
WITH cached_users AS (WITH cached_groups AS (select * from groups)
 SELECT uid, username, type, groupname, shell
 FROM users LEFT JOIN cached_groups USING (gid)
 WHERE type <> 'special' AND shell NOT LIKE '%/false' AND shell NOT LIKE '%/nologin' AND shell NOT LIKE '%/shutdown' AND shell NOT LIKE '%/halt' AND username NOT LIKE '%$' AND username NOT LIKE '\_%' ESCAPE '\' AND NOT (username = 'sync' AND shell ='/bin/sync' AND directory <> ''))
SELECT
  name AS name,
  bundle_short_version AS version,
  'Application (macOS)' AS type,
  bundle_identifier AS bundle_identifier,
  'apps' AS source,
  last_opened_time AS last_opened_at
FROM apps
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Python)' AS type,
  '' AS bundle_identifier,
  'python_packages' AS source,
  0 AS last_opened_at
FROM python_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Chrome)' AS type,
  '' AS bundle_identifier,
  'chrome_extensions' AS source,
  0 AS last_opened_at
FROM cached_users CROSS JOIN chrome_extensions USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Firefox)' AS type,
  '' AS bundle_identifier,
  'firefox_addons' AS source,
  0 AS last_opened_at
FROM cached_users CROSS JOIN firefox_addons USING (uid)
UNION
SELECT
  name As name,
  version AS version,
  'Browser plugin (Safari)' AS type,
  '' AS bundle_identifier,
  'safari_extensions' AS source,
  0 AS last_opened_at
FROM cached_users CROSS JOIN safari_extensions USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Atom)' AS type,
  '' AS bundle_identifier,
  'atom_packages' AS source,
  0 AS last_opened_at
FROM cached_users CROSS JOIN atom_packages USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Homebrew)' AS type,
  '' AS bundle_identifier,
  'homebrew_packages' AS source,
  0 AS last_opened_at
FROM homebrew_packages;
```

## software_windows

- Platforms: windows

- Query:

```sql
WITH cached_users AS (WITH cached_groups AS (select * from groups)
 SELECT uid, username, type, groupname, shell
 FROM users LEFT JOIN cached_groups USING (gid)
 WHERE type <> 'special' AND shell NOT LIKE '%/false' AND shell NOT LIKE '%/nologin' AND shell NOT LIKE '%/shutdown' AND shell NOT LIKE '%/halt' AND username NOT LIKE '%$' AND username NOT LIKE '\_%' ESCAPE '\' AND NOT (username = 'sync' AND shell ='/bin/sync' AND directory <> ''))
SELECT
  name AS name,
  version AS version,
  'Program (Windows)' AS type,
  'programs' AS source,
  publisher AS vendor
FROM programs
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Python)' AS type,
  'python_packages' AS source,
  '' AS vendor
FROM python_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (IE)' AS type,
  'ie_extensions' AS source,
  '' AS vendor
FROM ie_extensions
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Chrome)' AS type,
  'chrome_extensions' AS source,
  '' AS vendor
FROM cached_users CROSS JOIN chrome_extensions USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Browser plugin (Firefox)' AS type,
  'firefox_addons' AS source,
  '' AS vendor
FROM cached_users CROSS JOIN firefox_addons USING (uid)
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Chocolatey)' AS type,
  'chocolatey_packages' AS source,
  '' AS vendor
FROM chocolatey_packages
UNION
SELECT
  name AS name,
  version AS version,
  'Package (Atom)' AS type,
  'atom_packages' AS source,
  '' AS vendor
FROM cached_users CROSS JOIN atom_packages USING (uid);
```

## system_info

- Platforms: all

- Query:

```sql
select * from system_info limit 1
```

## uptime

- Platforms: all

- Query:

```sql
select * from uptime limit 1
```

## users

- Platforms: all

- Query:

```sql
WITH cached_groups AS (select * from groups)
 SELECT uid, username, type, groupname, shell
 FROM users LEFT JOIN cached_groups USING (gid)
 WHERE type <> 'special' AND shell NOT LIKE '%/false' AND shell NOT LIKE '%/nologin' AND shell NOT LIKE '%/shutdown' AND shell NOT LIKE '%/halt' AND username NOT LIKE '%$' AND username NOT LIKE '\_%' ESCAPE '\' AND NOT (username = 'sync' AND shell ='/bin/sync' AND directory <> '')
```

## windows_update_history

- Platforms: windows

- Discovery query:

```sql
SELECT 1 FROM osquery_registry WHERE active = true AND registry = 'table' AND name = 'windows_update_history';
```

- Query:

```sql
SELECT date, title FROM windows_update_history WHERE result_code = 'Succeeded'
```



<meta name="pageOrderInSection" value="1600">