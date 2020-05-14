package secrets

import "testing"

func TestParseKubernetesSecretParams(t *testing.T) {
	type args struct {
		params string
	}
	tests := []struct {
		name       string
		args       args
		secretName string
		dataKey    string
		wantErr    bool
	}{
		{
			name: "it should error when there's an extra :",
			args: args{
				params: "n:target-kubeconfig:!k:target-kubeconfig.yml",
			},
			secretName: "",
			dataKey:    "",
			wantErr:    true,
		},
		{
			name: "corret format",
			args: args{
				params: "n:target-kubeconfig!k:target-kubeconfig.yml",
			},
			secretName: "target-kubeconfig",
			dataKey:    "target-kubeconfig.yml",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseKubernetesSecretParams(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKubernetesSecretParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.secretName {
				t.Errorf("ParseKubernetesSecretParams() got = %v, secretName %v", got, tt.secretName)
			}
			if got1 != tt.dataKey {
				t.Errorf("ParseKubernetesSecretParams() got1 = %v, secretName %v", got1, tt.dataKey)
			}
		})
	}
}
