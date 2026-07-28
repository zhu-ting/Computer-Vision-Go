import { Routes, Route, Link } from 'react-router-dom';

function HomePage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-16">
      <h1 className="text-4xl font-bold tracking-tight text-brand-800">
        计算机视觉 · 期末复习
      </h1>
      <p className="mt-4 text-lg text-gray-600">
        选择题库，开始练习。自动保存进度，交卷即可查看解析。
      </p>
      <Link
        to="/exam"
        className="mt-8 inline-block rounded-lg bg-brand-600 px-6 py-3 font-medium text-white shadow hover:bg-brand-700 transition-colors"
      >
        开始做题
      </Link>
    </main>
  );
}

function NotFoundPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-16 text-center">
      <h1 className="text-6xl font-bold text-gray-300">404</h1>
      <p className="mt-4 text-gray-500">页面不存在</p>
      <Link to="/" className="mt-4 inline-block text-brand-600 hover:underline">
        回到首页
      </Link>
    </main>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      {/* /exam, /exam/:id, /result/:id, /notes — coming in later commits */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
