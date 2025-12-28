import ModuleDetail from '@/components/ModuleDetail'

export function generateStaticParams() {
  return [
    { moduleId: 'cluster-setup' },
    { moduleId: 'cluster-hardening' },
    { moduleId: 'system-hardening' },
    { moduleId: 'microservice-security' },
    { moduleId: 'supply-chain' },
    { moduleId: 'monitoring' },
  ]
}

export default function ModuleDetailPage({ params }: { params: { moduleId: string } }) {
  return <ModuleDetail moduleId={params.moduleId} />
}
