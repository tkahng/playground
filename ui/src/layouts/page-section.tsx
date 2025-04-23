import { Outlet } from "react-router";

export default function PageSectionLayout({ title }: { title?: string }) {
  return (
    <div className="flex-1">
      {title && (
        <header className="border-b w-full">
          <div className="flex-1">
            <div className="mx-auto px-12 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
              <h1 className="text-4xl font-bold text">{title}</h1>
            </div>
          </div>
        </header>
      )}
      <div className="mx-auto px-12 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
        <Outlet />
      </div>
    </div>
  );
}
