import { useState } from "react";
import { Sidebar } from "@/components/Sidebar";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { PanelLeft, Sun, Moon } from "lucide-react";
import { UploadReportModal } from "./UploadReportModal";
import { Switch } from "@/components/ui/switch";
import { useTheme } from "../hooks/useTheme";
import { Link } from 'react-router-dom';
import { AuthenticationButton } from "./AuthenticationButton";

interface DashboardLayoutProps {
  children: React.ReactNode;
  onUploadSuccess: () => void;
}

export function DashboardLayout({ children, onUploadSuccess }: DashboardLayoutProps) {
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const { theme, toggleTheme } = useTheme();

  const handleUploadReportClick = () => {
    setIsMobileMenuOpen(false);
    setIsUploadModalOpen(true);
  };

  return (
    <div className="h-screen flex">
      {/* Persistent Sidebar for Desktop */}
      <div className="hidden md:block fixed top-0 left-0 h-full w-[280px] border-r bg-muted/40">
        <Sidebar onUploadReportClick={handleUploadReportClick} />
      </div>

      <div className="flex flex-col flex-1 md:ml-[280px]">
        <header className="flex h-16 items-center justify-between border-b bg-background px-6 sticky top-0 z-10">
          {/* Mobile Navigation */}
          <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button size="icon" variant="outline" className="md:hidden">
                <PanelLeft className="h-5 w-5" />
                <span className="sr-only">Toggle Menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="p-0 w-64">
              <Sidebar onUploadReportClick={handleUploadReportClick} />
            </SheetContent>
          </Sheet>
          
          {/* Header Actions */}
          <div className="flex w-full items-center gap-4 justify-end">
            <Link to="/about" className="text-sm font-medium text-muted-foreground transition-colors hover:text-primary">
              About
            </Link>
            <div className="flex items-center space-x-2">
              {theme === "light" ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
              <Switch
                checked={theme === "dark"}
                onCheckedChange={toggleTheme}
              />
            </div>
            <AuthenticationButton />
          </div>
        </header>

        <main className="flex-1 flex flex-col overflow-y-auto">
          <div className="p-6 h-full">
            {children}
          </div>
        </main>
        <Sheet open={isUploadModalOpen} onOpenChange={setIsUploadModalOpen}>
              <SheetContent side="bottom" className="w-[280px]">
                <SheetHeader>
                  <SheetTitle>Upload Report</SheetTitle>
                  <SheetDescription>
                    Select a report type and upload your file.
                  </SheetDescription>
                </SheetHeader>
                <UploadReportModal onClose={() => setIsUploadModalOpen(false)} onUploadSuccess={onUploadSuccess} />
              </SheetContent>
            </Sheet>
      </div>
      
    </div>
  );
}
