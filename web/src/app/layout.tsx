import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'CKS Weight Room',
  description: 'Kubernetes Security Practice Environment',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
