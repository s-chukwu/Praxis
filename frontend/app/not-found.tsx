import ErrorScreen from '@/components/ErrorScreen';

export default function NotFound() {
  return <ErrorScreen title="Page not found" message="This page does not exist or has moved." />;
}
