/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Brand palette — clean, academic feel
        brand: {
          50:  '#f0fdf9', // 极淡的薄荷冷白，适合页面底色或悬停卡片
          100: '#ccfbef', // 浅色背景高亮、标签背景
          200: '#99f6e0', // 边框、分割线
          300: '#5eead4', // 次级强调色
          400: '#2dd4bf', // 图标高亮、明亮装饰元素
          500: '#0d9488', // 基准品牌色（按钮、主行动点）
          600: '#0f766e', // 悬停态、次级标题
          700: '#115e59', // 常用正文强调、深色按钮
          800: '#134e48', // 对应原 800：沉稳深邃的翡翠墨绿，适合大标题和高对比文本
          900: '#042f2e', // 最深沉的森林墨色，适合深色侧边栏或主文本
        },
      },
    },
  },
  plugins: [],
};
