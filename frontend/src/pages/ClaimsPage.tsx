// frontend/src/pages/ClaimsPage.tsx

import { useEffect, useState } from "react";
import { DataTable } from "@/components/ui/DataTable";
import { columns, Claim } from "@/components/claims/columns"; // We'll create this next
import { DetailsDrawer } from "@/components/DetailsDrawer";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { useAuth0 } from "@auth0/auth0-react";
import { apiClient } from "@/lib/api";

export default function ClaimsPage() {
  const [claims, setClaims] = useState<Claim[]>([]);
  const [selectedClaim, setSelectedClaim] = useState<Claim | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { getAccessTokenSilently } = useAuth0();

  // --- Data Fetching Effect ---
  useEffect(() => {
    const fetchClaims = async () => {
      setIsLoading(true);
      try {
        const token = await getAccessTokenSilently();
        const data = await apiClient.get('/api/insurance/claims', token);
        setClaims(data || []); // Handle null response from API
      } catch (error) {
        console.error("Failed to fetch claims:", error);
        // Here you would normally set an error state and show a toast
      } finally {
        setIsLoading(false);
      }
    };
    fetchClaims();
  }, [getAccessTokenSilently]);

  // --- Event Handlers ---
  const handleRowClick = (claim: Claim) => {
    setSelectedClaim(claim);
  };

  const handleDrawerClose = () => {
    setSelectedClaim(null);
  };
  
  const handleSave = (updatedClaim: Claim) => {
    // This is where we'll add the API call to update a claim later
    console.log("Saving claim:", updatedClaim);
    // Optimistically update the UI
    setClaims(prevClaims => prevClaims.map(c => c.id === updatedClaim.id ? updatedClaim : c));
    setSelectedClaim(null); // Close the drawer on save
  };

  if (isLoading) {
    return <div>Loading claims data...</div>;
  }

  return (
    <>
      <DataTable
        columns={columns}
        data={claims}
        title="Insurance Claims"
        description="Browse and manage all insurance claims."
        page={1} // We will wire up pagination later
        setPage={() => {}}
        hasMore={false}
        onRowClick={handleRowClick}
      />

      <Sheet open={!!selectedClaim} onOpenChange={(isOpen) => !isOpen && handleDrawerClose()}>
        <SheetContent className="sm:max-w-2xl">
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
                main: [],
                status: [
                  {key: 'business_status', label: 'Status', options: ['Submitted', 'Under Review', 'Flagged for Fraud Review', 'Approved', 'Paid', 'Denied'] },
                ],
                comments: [] 
              }} // We'll define these fields next
              onSave={handleSave}
              onCancel={handleDrawerClose}
              id={selectedClaim.id}
              type="item"
            />
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
