// frontend/src/pages/ClaimsPage.tsx

import { useEffect, useState } from "react";
import { DataTable } from "@/components/ui/DataTable";
import { columns, Claim } from "@/components/claims/columns";
import { DetailsDrawer } from "@/components/DetailsDrawer";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { useAuth0 } from "@auth0/auth0-react";
import { apiClient } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"; // Import Card for the chat
import toast from "react-hot-toast";

export default function ClaimsPage() {
  const [claims, setClaims] = useState<Claim[]>([]);
  const [selectedClaim, setSelectedClaim] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { getAccessTokenSilently } = useAuth0();

  // --- Data fetching and handlers remain the same ---
  useEffect(() => {
    const fetchClaims = async () => {
      setIsLoading(true);
      try {
        const token = await getAccessTokenSilently();
        const data = await apiClient.get('/api/insurance/claims', token);
        setClaims(data || []);
      } catch (error) {
        console.error("Failed to fetch claims:", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchClaims();
  }, [getAccessTokenSilently]);
  
  const fetchClaimDetails = async (itemId: number) => {
    try {
      const token = await getAccessTokenSilently();
      const data = await apiClient.get(`/api/insurance/claims/${itemId}`, token);
      setSelectedClaim(data);
    } catch (error) {
      console.error(`Failed to fetch details for claim id ${itemId}:`, error);
    }
  };

  const handleRowClick = (claim: Claim) => {
    fetchClaimDetails(claim.id); 
  };

  const handleDrawerClose = () => {
    setSelectedClaim(null);
  };
  
  const handleSave = async (updatedClaim: any) => {
    const originalClaims = [...claims];

    setClaims(prevClaims => prevClaims.map(c => c.id === updatedClaim.id ? { ...c, business_status: updatedClaim.business_status } : c));
    setSelectedClaim(null);

    const toastId = toast.loading("Saving status update...");

    try {
      const token = await getAccessTokenSilently();
      const payload = { business_status: updatedClaim.business_status };

      await apiClient.patch(`/api/insurance/claims/${updatedClaim.id}`, token, payload);

      toast.success("Status updated successfully!", { id: toastId });

    } catch (error) {
      toast.error("Failed to save status update.", { id: toastId });
      console.error("Failed to save claim:", error);
      setClaims(originalClaims);
    }
};

  if (isLoading) {
    return <div className="h-full flex items-center justify-center">Loading claims data...</div>;
  }

  return (
    <div className="h-full flex flex-row gap-6">
      
      {/* --- Column 1: AI Chat --- */}
      <div className="w-[30%] flex-shrink-0">
        <Card className="h-full">
          <CardHeader>
            <CardTitle>AI Assistant</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">Chat interface will go here.</p>
          </CardContent>
        </Card>
      </div>

      {/* --- Column 2: Data Table --- */}
      <div className="w-[70%] flex flex-col min-w-0">
        <DataTable
          columns={columns}
          data={claims}
          title="General Securities Assurance - Policy Claims"
          description="Browse and manage all insurance claims."
          page={1}
          setPage={() => {}}
          hasMore={false}
          onRowClick={handleRowClick}
        />
      </div>

      {/* --- Details Drawer (slides from the right) --- */}
      <Sheet open={!!selectedClaim} onOpenChange={(isOpen) => !isOpen && handleDrawerClose()}>
        {/* The side="right" prop is the key change here */}
        <SheetContent side="right" className="sm:max-w-2xl">
          <SheetHeader>
            <SheetTitle>Claim Details: {selectedClaim?.claim_id}</SheetTitle>
            <SheetDescription>
              View and edit the details for the selected claim.
            </SheetDescription>
          </SheetHeader>
          {selectedClaim && (
            <DetailsDrawer
              data={selectedClaim}
              fields={{
                main: [
                  { key: 'policy_number', label: 'Policy #' },
                  { key: 'claim_type', label: 'Claim Type' },
                  { key: 'date_of_loss', label: 'Date of Loss' },
                  { key: 'claim_amount', label: 'Claim Amount', type: 'currency' },
                  { key: 'adjuster_assigned', label: 'Adjuster' },
                  { key: 'policyholder_name', label: 'Policyholder' },
                  { key: 'customer_level', label: 'Customer Level' },
                  { key: 'customer_since_date', label: 'Customer Since' },
                ],
                status: [
                  { key: 'business_status', label: 'Status', options: ['Submitted', 'Under Review', 'Flagged for Fraud Review', 'Approved', 'Paid', 'Denied'] },
                ],
                comments: [],
              }}
              onSave={handleSave}
              onCancel={handleDrawerClose}
              id={selectedClaim.id}
              type="item"
            />
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
