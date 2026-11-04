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
Render-time validation shared by the chart templates. Called from the
ConfigMap, which is always rendered.
*/}}
{{- define "cubeface.validate" -}}
{{- $mode := toString (default "sqlite" .Values.database.mode) -}}
{{- if not (has $mode (list "sqlite" "postgres")) -}}
{{- fail (printf "database.mode must be 'sqlite' or 'postgres', got %q" $mode) -}}
{{- end -}}

{{- $replicas := int (default 1 .Values.replicaCount) -}}
{{- if lt $replicas 1 -}}
{{- fail "replicaCount must be at least 1" -}}
{{- end -}}
{{- if and (eq $mode "sqlite") (gt $replicas 1) -}}
{{- fail "replicaCount > 1 requires database.mode=postgres: SQLite is a single-writer file and cannot be shared between pods" -}}
{{- end -}}

{{- if and (eq $mode "postgres") .Values.secret.create -}}
{{- if not (trim (toString .Values.database.postgres.dsn)) -}}
{{- fail "database.postgres.dsn is required when database.mode=postgres and secret.create=true (or set secret.create=false and provide SQL_DSN in secret.existingSecret)" -}}
{{- end -}}
{{- end -}}

{{- $_ := include "cubeface.secretName" . -}}
{{- $_ := include "cubeface.basePath" . -}}

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
