{{/*
Chart name. Kept independent of .Chart.Name because it is used in immutable
Deployment selectors.
*/}}
{{- define "cubeface.name" -}}
{{- default "cubeface" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Fully qualified app name, used as the base for every resource name.
*/}}
{{- define "cubeface.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else if .Values.nameOverride -}}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "cubeface.labels" -}}
app.kubernetes.io/name: {{ include "cubeface.name" . }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end -}}

{{/*
Selector labels for a component.
Usage: {{ include "cubeface.selectorLabels" (dict "root" . "component" "app") }}
*/}}
{{- define "cubeface.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cubeface.name" .root }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{/*
Normalized BASE_PATH: "", or "/foo" without a trailing slash.
*/}}
{{- define "cubeface.basePath" -}}
{{- $raw := trim (toString .Values.config.BASE_PATH) -}}
{{- $clean := trimAll "/" $raw -}}
{{- if $clean -}}
{{- if not (regexMatch "^[A-Za-z0-9._~-]+(/[A-Za-z0-9._~-]+)*$" $clean) -}}
{{- fail (printf "config.BASE_PATH must be a path prefix such as /cubeface, got %q" $raw) -}}
{{- end -}}
{{- range $segment := splitList "/" $clean -}}
{{- if or (eq $segment ".") (eq $segment "..") -}}
{{- fail (printf "config.BASE_PATH must not contain . or .. path segments, got %q" $raw) -}}
{{- end -}}
{{- end -}}
{{- printf "/%s" $clean -}}
{{- end -}}
{{- end -}}

{{/*
Browser-visible origin. Configured value wins; otherwise derive it from the
first Ingress host (https when TLS is enabled), falling back to the port-forward
address. Login POSTs are origin-checked, so this must match the browser URL.
*/}}
{{- define "cubeface.publicOrigin" -}}
{{- $configured := trim (toString .Values.config.PUBLIC_ORIGIN) -}}
{{- if $configured -}}
{{- trimSuffix "/" $configured -}}
{{- else if and .Values.ingress.enabled .Values.ingress.hosts -}}
{{- $scheme := ternary "https" "http" .Values.ingress.tls.enabled -}}
{{- printf "%s://%s" $scheme (first .Values.ingress.hosts) -}}
{{- else -}}
{{- printf "http://localhost:%s" (toString .Values.config.PORT) -}}
{{- end -}}
{{- end -}}

{{/*
Session cookie Secure flag: explicit value wins, otherwise https origins get
Secure cookies and plain-http local port-forwards remain usable.
*/}}
{{- define "cubeface.sessionCookieSecure" -}}
{{- $secure := "" -}}
{{- if hasKey .Values.config "SESSION_COOKIE_SECURE" -}}
{{- $secure = toString .Values.config.SESSION_COOKIE_SECURE -}}
{{- end -}}
{{- if $secure -}}
{{- if not (has $secure (list "true" "false")) -}}
{{- fail (printf "config.SESSION_COOKIE_SECURE must be \"true\", \"false\" or empty (auto), got %q" $secure) -}}
{{- end -}}
{{- $secure -}}
{{- else -}}
{{- if hasPrefix "https://" (include "cubeface.publicOrigin" .) -}}true{{- else -}}false{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Ingress path and probe path (the Go app serves /healthz under BASE_PATH).
*/}}
{{- define "cubeface.ingressPath" -}}
{{- $base := include "cubeface.basePath" . -}}
{{- if $base -}}{{ $base }}{{- else -}}/{{- end -}}
{{- end -}}

{{- define "cubeface.probePath" -}}
{{- printf "%s/healthz" (include "cubeface.basePath" .) -}}
{{- end -}}

{{/*
Created or pre-existing Secret name.
*/}}
{{- define "cubeface.secretName" -}}
{{- if .Values.secret.create -}}
{{- printf "%s-secret" (include "cubeface.fullname" .) -}}
{{- else -}}
{{- required "secret.existingSecret is required when secret.create=false" .Values.secret.existingSecret -}}
{{- end -}}
{{- end -}}

{{- define "cubeface.pvcName" -}}
{{- printf "%s-data" (include "cubeface.fullname" .) -}}
{{- end -}}

{{/*
CloudNativePG Cluster CR name. Truncated so the operator-derived service names
(<name>-rw, <name>-ro, <name>-pgbouncer) stay within 63 characters.
*/}}
{{- define "cubeface.postgresClusterName" -}}
{{- printf "%s-postgres" (include "cubeface.fullname" .) | trunc 50 | trimSuffix "-" -}}
{{- end -}}

{{/*
Writable PostgreSQL endpoint: the CloudNativePG read-write service follows the
current primary across failovers.
*/}}
{{- define "cubeface.postgresAddr" -}}
{{- printf "%s-rw:5432" (include "cubeface.postgresClusterName" .) -}}
{{- end -}}

{{/*
basic-auth Secret passed to spec.bootstrap.initdb.secret. The operator applies
its password to the database owner role and re-applies it on change.
*/}}
{{- define "cubeface.postgresAppAuthSecretName" -}}
{{- printf "%s-app-auth" (include "cubeface.postgresClusterName" .) -}}
{{- end -}}

{{/*
Chart-managed Secret holding SQL_DSN for the managed cluster.
*/}}
{{- define "cubeface.postgresDSNSecretName" -}}
{{- printf "%s-postgres-dsn" (include "cubeface.fullname" .) -}}
{{- end -}}

{{/*
Fail the render when a value embedded verbatim in a URI connection string
would break the userinfo segment. Generated passwords are alphanumeric and
always pass; explicit values with URI special characters are rejected instead
of being percent-encoded so the rendered DSN stays predictable for psql and
pgx alike.
Usage: {{ include "cubeface.assertDSNSafe" (dict "value" $pw "label" "secrets.POSTGRES_PASSWORD") }}
*/}}
{{- define "cubeface.assertDSNSafe" -}}
{{- $v := toString .value -}}
{{- if or (contains "'" $v) (contains " " $v) (contains "@" $v) (contains ":" $v) (contains "/" $v) (contains "?" $v) (contains "#" $v) (contains "[" $v) (contains "]" $v) (contains "%" $v) -}}
{{- fail (printf "%s must not contain single quotes, spaces or URI special characters (@ : / ? # [ ] %%)" .label) -}}
{{- end -}}
{{- end -}}

{{/*
Render-time validation shared by the chart templates. Called from the
ConfigMap, which is always rendered.
*/}}
{{- define "cubeface.validate" -}}
{{- $mode := toString (default "postgres" .Values.database.mode) -}}
{{- if not (has $mode (list "postgres" "external" "sqlite")) -}}
{{- fail (printf "database.mode must be 'postgres', 'external' or 'sqlite', got %q" $mode) -}}
{{- end -}}

{{- $replicas := int (default 1 .Values.replicaCount) -}}
{{- if lt $replicas 1 -}}
{{- fail "replicaCount must be at least 1" -}}
{{- end -}}
{{- if and (eq $mode "sqlite") (gt $replicas 1) -}}
{{- fail "replicaCount > 1 requires database.mode=postgres or external: SQLite is a single-writer file and cannot be shared between pods" -}}
{{- end -}}

{{- if eq $mode "external" -}}
{{- if and .Values.secret.create (not (trim (toString .Values.database.external.dsn))) -}}
{{- fail "database.external.dsn is required when database.mode=external and secret.create=true (or set secret.create=false and provide SQL_DSN in secret.existingSecret)" -}}
{{- end -}}
{{- end -}}

{{- if eq $mode "postgres" -}}
{{- $pg := .Values.database.postgres -}}
{{- if not (trim (toString $pg.image)) -}}
{{- fail "database.postgres.image is required when database.mode=postgres" -}}
{{- end -}}
{{- if lt (int (default 0 $pg.instances)) 1 -}}
{{- fail "database.postgres.instances must be at least 1" -}}
{{- end -}}
{{- if not (trim (toString $pg.storage)) -}}
{{- fail "database.postgres.storage is required when database.mode=postgres" -}}
{{- end -}}
{{- $pgUser := trim (toString $pg.auth.username) -}}
{{- if not (regexMatch "^[a-z_][a-z0-9_]*$" $pgUser) -}}
{{- fail (printf "database.postgres.auth.username must match [a-z_][a-z0-9_]* (a PostgreSQL identifier embedded in SQL_DSN), got %q" $pgUser) -}}
{{- end -}}
{{- $pgDatabase := trim (toString $pg.auth.database) -}}
{{- if not (regexMatch "^[a-z_][a-z0-9_]*$" $pgDatabase) -}}
{{- fail (printf "database.postgres.auth.database must match [a-z_][a-z0-9_]* (a PostgreSQL identifier embedded in SQL_DSN), got %q" $pgDatabase) -}}
{{- end -}}
{{- $pgPassword := toString (default "" .Values.secrets.POSTGRES_PASSWORD) -}}
{{- if $pgPassword -}}
{{- include "cubeface.assertDSNSafe" (dict "value" $pgPassword "label" "secrets.POSTGRES_PASSWORD") -}}
{{- end -}}
{{- if and $pg.pgBouncer.enabled (lt (int (default 0 $pg.pgBouncer.replicas)) 1) -}}
{{- fail "database.postgres.pgBouncer.replicas must be at least 1 when the pooler is enabled" -}}
{{- end -}}
{{- if $pg.backups.enabled -}}
{{- $method := toString $pg.backups.method -}}
{{- if not (has $method (list "volumeSnapshot" "fileArchiving")) -}}
{{- fail (printf "database.postgres.backups.method must be volumeSnapshot or fileArchiving, got %q" $method) -}}
{{- end -}}
{{- if and (eq $method "fileArchiving") (not $pg.backups.barmanObjectStore) -}}
{{- fail "database.postgres.backups.barmanObjectStore is required when backups.method=fileArchiving" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- $_ := include "cubeface.secretName" . -}}
{{- $_ := include "cubeface.basePath" . -}}
{{- if eq $mode "postgres" -}}
{{- $_ := include "cubeface.postgresClusterName" . -}}
{{- end -}}

{{- $origin := include "cubeface.publicOrigin" . -}}
{{- if not (regexMatch "^https?://[^/]+$" $origin) -}}
{{- fail (printf "config.PUBLIC_ORIGIN must be an http(s) origin without a path, got %q" $origin) -}}
{{- end -}}

{{- $port := int (toString .Values.config.PORT) -}}
{{- if or (lt $port 1) (gt $port 65535) -}}
{{- fail (printf "config.PORT must be a TCP port, got %q" (toString .Values.config.PORT)) -}}
{{- end -}}

{{- $bools := dict "REGISTER_ENABLED" .Values.config.REGISTER_ENABLED "EMAIL_VERIFICATION_ENABLED" .Values.config.EMAIL_VERIFICATION_ENABLED "EMAIL_DEBUG_MODE" .Values.config.EMAIL_DEBUG_MODE "ALLOW_EMAIL_DEBUG_IN_PRODUCTION" .Values.config.ALLOW_EMAIL_DEBUG_IN_PRODUCTION -}}
{{- range $key, $value := $bools -}}
{{- $text := toString $value -}}
{{- if not (has $text (list "true" "false")) -}}
{{- fail (printf "config.%s must be \"true\" or \"false\", got %q" $key $text) -}}
{{- end -}}
{{- end -}}

{{- $tlsMode := toString .Values.config.SMTP_TLS_MODE -}}
{{- if not (has $tlsMode (list "implicit" "starttls" "none")) -}}
{{- fail (printf "config.SMTP_TLS_MODE must be implicit, starttls or none, got %q" $tlsMode) -}}
{{- end -}}

{{- if eq (toString .Values.config.EMAIL_VERIFICATION_ENABLED) "true" -}}
{{- if eq (toString .Values.config.EMAIL_DEBUG_MODE) "true" -}}
{{- if ne (toString .Values.config.ALLOW_EMAIL_DEBUG_IN_PRODUCTION) "true" -}}
{{- fail "EMAIL_DEBUG_MODE=true in production requires config.ALLOW_EMAIL_DEBUG_IN_PRODUCTION=true; prefer configuring SMTP instead" -}}
{{- end -}}
{{- else if not (trim (toString .Values.config.SMTP_HOST)) -}}
{{- fail "config.SMTP_HOST is required when config.EMAIL_VERIFICATION_ENABLED=true (or disable email verification)" -}}
{{- end -}}
{{- end -}}

{{- if .Values.ingress.enabled -}}
{{- if not .Values.ingress.hosts -}}
{{- fail "ingress.hosts must contain at least one host when ingress.enabled=true" -}}
{{- end -}}
{{- if and .Values.ingress.tls.enabled (not (trim (toString .Values.ingress.tls.secretName))) -}}
{{- fail "ingress.tls.secretName is required when ingress.tls.enabled=true" -}}
{{- end -}}
{{- end -}}

{{- $reserved := list "APP_ENV" "PORT" "BASE_PATH" "PUBLIC_ORIGIN" "SYSTEM_NAME" "REGISTER_ENABLED" "EMAIL_VERIFICATION_ENABLED" "EMAIL_DEBUG_MODE" "ALLOW_EMAIL_DEBUG_IN_PRODUCTION" "SESSION_COOKIE_SECURE" "SESSION_TTL_DAYS" "SESSION_TOUCH_INTERVAL_SECONDS" "TRUSTED_PROXIES" "SMTP_HOST" "SMTP_PORT" "SMTP_TLS_MODE" "SMTP_FROM" "SMTP_USERNAME" "GITHUB_CLIENT_ID" "GOOGLE_CLIENT_ID" "OIDC_ISSUER_URL" "OIDC_CLIENT_ID" "OIDC_DISPLAY_NAME" "OIDC_SCOPES" "INITIAL_ADMIN_USERNAME" "INITIAL_ADMIN_EMAIL" "SQLITE_PATH" "SQL_DSN" "SESSION_SECRET" -}}
{{- range $key, $_ := (default (dict) .Values.config.extra) -}}
{{- if has $key $reserved -}}
{{- fail (printf "config.extra.%s conflicts with a key managed by the chart; set the top-level value instead" $key) -}}
{{- end -}}
{{- end -}}
{{- end -}}
