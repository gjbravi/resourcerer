{{- define "resourcerer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "resourcerer.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "resourcerer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "resourcerer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "resourcerer.name" .root }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{- define "resourcerer.labels" -}}
helm.sh/chart: {{ include "resourcerer.chart" . }}
app.kubernetes.io/name: {{ include "resourcerer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "resourcerer.excludedNamespaces" -}}
{{- $ns := concat .Values.excludedNamespaces (list .Release.Namespace) | uniq }}
{{- join "," $ns }}
{{- end }}

{{- /* Flags every component shares. */ -}}
{{- define "resourcerer.commonArgs" -}}
- --cluster-name={{ .Values.clusterName }}
- --selection={{ .Values.selection }}
- --excluded-namespaces={{ include "resourcerer.excludedNamespaces" . }}
- --defaults-configmap={{ .Release.Namespace }}/{{ include "resourcerer.fullname" . }}-defaults
- --prometheus-url={{ .Values.prometheus.url }}
- --rate-window={{ .Values.prometheus.rateWindow }}
- --min-step={{ .Values.prometheus.minStep }}
- --health-probe-bind-address=:8082
{{- end }}

{{- /* Deployment for one component. Call with dict "root" $ "component" name "values" .Values.<component> "args" list "ports" list. */ -}}
{{- define "resourcerer.deployment" -}}
{{- $root := .root }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "resourcerer.fullname" $root }}-{{ .component }}
  namespace: {{ $root.Release.Namespace }}
  labels:
    {{- include "resourcerer.labels" $root | nindent 4 }}
    app.kubernetes.io/component: {{ .component }}
spec:
  replicas: {{ .values.replicas }}
  selector:
    matchLabels:
      {{- include "resourcerer.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        checksum/defaults: {{ include (print $root.Template.BasePath "/configmap.yaml") $root | sha256sum | trunc 16 }}
        {{- with $root.Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      labels:
        {{- include "resourcerer.labels" $root | nindent 8 }}
        app.kubernetes.io/component: {{ .component }}
        {{- with $root.Values.podLabels }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
    spec:
      serviceAccountName: {{ include "resourcerer.fullname" $root }}-{{ .component }}
      {{- with $root.Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.priorityClassName }}
      priorityClassName: {{ . }}
      {{- end }}
      securityContext:
        {{- toYaml $root.Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .component }}
          image: "{{ $root.Values.image.repository }}:{{ $root.Values.image.tag | default $root.Chart.AppVersion }}"
          imagePullPolicy: {{ $root.Values.image.pullPolicy }}
          args:
            - {{ .component }}
            {{- include "resourcerer.commonArgs" $root | nindent 12 }}
            {{- if gt (int .values.replicas) 1 }}
            - --leader-elect
            {{- end }}
            {{- range .args }}
            - {{ . }}
            {{- end }}
            {{- range .values.extraArgs }}
            - {{ . }}
            {{- end }}
          ports:
            {{- range .ports }}
            - name: {{ .name }}
              containerPort: {{ .port }}
            {{- end }}
            - name: probes
              containerPort: 8082
          livenessProbe:
            httpGet: { path: /healthz, port: probes }
          readinessProbe:
            httpGet: { path: /readyz, port: probes }
          securityContext:
            {{- toYaml $root.Values.securityContext | nindent 12 }}
          resources:
            {{- toYaml .values.resources | nindent 12 }}
      {{- with $root.Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
