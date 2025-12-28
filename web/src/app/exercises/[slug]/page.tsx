import ExerciseClientPage from './ExerciseClientPage'

// Generate static paths for all 20 exercises
export function generateStaticParams() {
  return [
    { slug: 'falco-dev-mem-detection' },
    { slug: 'bom-libcrypto-version' },
    { slug: 'docker-group-tcp-hardening' },
    { slug: 'projected-volume-sa-token' },
    { slug: 'imagepolicywebhook-admission' },
    { slug: 'audit-policy-configuration' },
    { slug: 'networkpolicy-default-deny' },
    { slug: 'ingress-tls-redirect' },
    { slug: 'kube-bench-cis-fixes' },
    { slug: 'cilium-network-policy-mtls' },
    { slug: 'trivy-image-scan' },
    { slug: 'kubeadm-node-upgrade' },
    { slug: 'pod-security-standards' },
    { slug: 'istio-sidecar-mtls' },
    { slug: 'gvisor-runtime-class' },
    { slug: 'static-analysis-security' },
    { slug: 'etcd-encryption-at-rest' },
    { slug: 'disable-anonymous-access' },
    { slug: 'container-immutability' },
    { slug: 'verify-platform-binaries' },
  ]
}

export default function ExerciseDetailPage({ params }: { params: { slug: string } }) {
  return <ExerciseClientPage slug={params.slug} />
}
