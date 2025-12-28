import PracticeView from './PracticeView'

// Generate static paths for all exercises
export function generateStaticParams() {
  return [
    { slug: 'falco-dev-mem-detection' },
    { slug: 'bom-libcrypto-version' },
    { slug: 'docker-tls-version' },
    { slug: 'trivy-image-scan' },
    { slug: 'seccomp-profile' },
    { slug: 'apparmor-profile' },
    { slug: 'audit-policy-configuration' },
    { slug: 'tls-certificate-rotation' },
    { slug: 'disable-anonymous-access' },
    { slug: 'rbac-least-privilege' },
    { slug: 'networkpolicy-default-deny' },
    { slug: 'networkpolicy-egress-dns' },
    { slug: 'projected-volume-sa-token' },
    { slug: 'pod-security-standard' },
    { slug: 'imagepolicywebhook-admission' },
    { slug: 'disable-automount-sa-token' },
    { slug: 'verify-binaries-checksum' },
    { slug: 'etcd-encryption-at-rest' },
    { slug: 'static-analysis-kubesec' },
    { slug: 'runtime-detection-falco' },
  ]
}

export default function PracticePage({ params }: { params: { slug: string } }) {
  return <PracticeView slug={params.slug} />
}
