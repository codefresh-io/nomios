{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "nomios.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "nomios.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "nomios.fullnameOverride" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- $namePrefix := default .Release.Name .Values.global.releaseNameOverride -}}
{{- printf "%s-%s" $namePrefix $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app role.
*/}}
{{- define "nomios.role" -}}
{{- default "trigger-dockerhub" .Values.roleOverride -}}
{{- end -}}

{{/*
Configure public DNS name for constructing webhook url
first, look for `global.appURL`, if not set fallback to `https://g.codefresh.io`
*/}}
{{- define "nomios.publicDNS" -}}
{{- if .Values.global -}}
{{- if .Values.global.appUrl -}}
{{- .Values.global.appUrl -}}
{{- else -}}
{{- default .Values.publicDnsName "g.codefresh.io" -}}
{{- end -}}
{{- else -}}
{{- default .Values.publicDnsName "g.codefresh.io" -}}
{{- end -}}
{{- end -}}

{{/*
   Create a default fully qualified app name.
   We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
   If release name contains chart name it will be used as a full name.
  */}}
{{- define "nomios.fqdn" -}}
  {{- $name := "" -}}
  {{- if $.Values.fullnameOverride -}}
    {{- $name =$.Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
  {{- else -}}
    {{- $name = default .Chart.Name .Values.nameOverride -}}
    {{- if contains $name .Release.Name -}}
      {{- $name = .Release.Name | trunc 63 | trimSuffix "-" -}}
    {{- else -}}
      {{- $name = printf "%s-%s" .Release.Name $name -}}
    {{- end -}}
  {{- end -}}
  {{- printf "%s.%s.svc.cluster.local" $name .Release.Namespace  | trunc 63 | trimSuffix "-" -}}
{{- end -}}