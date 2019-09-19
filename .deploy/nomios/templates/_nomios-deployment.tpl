{{/*
We create Deployment resource as template to be able to use many deployments but with 
different name and version. This is for Istio POC.
*/}}
{{- define "nomios.renderDeployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ template "nomios.fullname" $ }}-{{ .version | default "base" }}
  labels:
    app: {{ template "nomios.name" . }}
    role: {{ template "nomios.role" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
    version: {{ .version | default "base" | quote  }}
    stable-version: {{ .stableVersion | default "false" | quote  }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ template "nomios.name" . }}
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print .Template.BasePath "/configmap.yaml") . | sha256sum }}
        sidecar.istio.io/inject: {{ or $.Values.istio.enabled $.Values.global.istio.enabled  | default false | quote }}
      labels:
        app: {{ template "nomios.name" . }}
        role: {{ template "nomios.role" . }}
        release: {{ .Release.Name }}
        type: {{ .Values.event.type }}
        kind: {{ .Values.event.kind }}
        action: {{ .Values.event.action }}
        version: {{ .version | default "base" | quote  }}
        stable-version: {{ .stableVersion | default "false" | quote  }}
    spec:
      imagePullSecrets:
        - name: "{{ .Release.Name }}-{{ .Values.global.codefresh }}-registry"
      containers:
        - name: {{ .Chart.Name }}
          {{- if .Values.global.privateRegistry }}
          image: "{{ .Values.global.dockerRegistry }}{{ .Values.image.name }}:{{ .imageTag }}"
          {{- else }}
          image: "{{ .Values.image.dockerRegistry }}{{ .Values.image.name }}:{{ .imageTag }}"
          {{- end }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: {{ .Values.service.internalPort }}
          env:
            {{- if .Values.global.env }}
              {{- range $key, $value := .Values.global.env }}
              - name: {{ $key }}
                value: {{ $value | quote }}
              {{- end}}
              {{- end}}
              - name: LOG_LEVEL
                value: {{ .Values.logLevel | quote }}
              {{- if ne .Values.logLevel "debug" }}
              - name: GIN_MODE
                value: release
              {{- end }}
              - name: HERMES_SERVICE
                value: {{ .Values.hermesService | default (printf "%s-hermes" .Release.Name) }}
              - name: PUBLIC_DNS_NAME
                value: "https://{{ template "nomios.publicDNS" . }}"
              - name: PORT
                value: {{ .Values.service.internalPort | quote }}
          livenessProbe:
            httpGet:
              path: /health
              port: {{ .Values.service.internalPort }}
            initialDelaySeconds: 5
            timeoutSeconds: 3
            periodSeconds: 15
            failureThreshold: 5
          readinessProbe:
            httpGet:
              path: /ping
              port: {{ .Values.service.internalPort }}
            initialDelaySeconds: 5
            timeoutSeconds: 3
            periodSeconds: 15
            failureThreshold: 5
          resources:
{{ toYaml .Values.resources | indent 12 }}
    {{- if .Values.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.nodeSelector | indent 8 }}
    {{- end }}
{{- end }}
