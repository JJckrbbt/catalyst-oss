import { useEffect, useState } from "react";
import { DataTable } from "@/components/ui/DataTable";
import { columns, Upload } from "@/components/uploads/columns";

async function getData(): Promise<Upload[]> {
  // Fetch data from your API here.
  // This is just a placeholder.
  return [
    {
      id: "728ed52f-7b29-4ca4-a5d2-9c5b9a3f3a2f",
      status: "Success",
      report_type: "items",
      uploadedAt: new Date().toISOString(),
      error_details: null,
      user: {
        first_name: "John",
        last_name: "Doe",
      },
    },
  ];
}

export default function UploadsPage() {
  const [data, setData] = useState<Upload[]>([]);
  const [page, setPage] = useState(1);

  useEffect(() => {
    getData().then(setData);
  }, []);

  return (
    <DataTable
      columns={columns}
      data={data}
      title="Uploads"
      description="View the status of your recent uploads."
      page={page}
      setPage={setPage}
      hasMore={false}
    />
  );
};
