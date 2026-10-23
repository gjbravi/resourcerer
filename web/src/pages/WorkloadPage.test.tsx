import { ThemeProvider } from "@mui/material/styles";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { workload } from "../test/fixtures";
import { theme } from "../theme";
import { WorkloadPage } from "./WorkloadPage";

function respond(routes: Record<string, unknown>) {
  return vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
    const url = String(input);
    const key = `${init?.method ?? "GET"} ${url}`;
    const body = routes[key] ?? routes[url];
    if (body === undefined) return new Response("{}", { status: 404 });
    return new Response(JSON.stringify(body), { status: 200 });
  });
}

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <ThemeProvider theme={theme}>
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={["/workloads/shop/Deployment/api"]}>
          <Routes>
            <Route path="/workloads/:namespace/:kind/:name" element={<WorkloadPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </ThemeProvider>,
  );
}

afterEach(() => vi.restoreAllMocks());

const info = { version: "test", selection: "all", defaultMode: "Off", readOnly: false };

describe("WorkloadPage", () => {
  it("shows the VPA recommendation", async () => {
    respond({ "/api/v1/info": info, "/api/v1/workloads/shop/Deployment/api": workload() });
    renderPage();
    expect(await screen.findByText("→ 126m")).toBeInTheDocument();
    expect(screen.getByText("→ 74Mi")).toBeInTheDocument();
    expect(screen.getByText("50m – 1500m")).toBeInTheDocument();
  });

  it("asks for confirmation before switching to Auto", async () => {
    const fetchMock = respond({
      "/api/v1/info": info,
      "/api/v1/workloads/shop/Deployment/api": workload(),
      "PUT /api/v1/workloads/shop/Deployment/api/mode": workload({ mode: "Auto", modeSource: "workload" }),
    });
    renderPage();
    await userEvent.click(await screen.findByRole("radio", { name: /^auto/i }));
    expect(await screen.findByRole("dialog")).toHaveTextContent(/start changing resources of running pods/i);
    expect(fetchMock.mock.calls.some(([, i]) => i?.method === "PUT")).toBe(false);

    await userEvent.click(screen.getByRole("button", { name: "Switch" }));
    await waitFor(() => {
      const put = fetchMock.mock.calls.find(([, i]) => i?.method === "PUT");
      expect(put && JSON.parse(put[1]!.body as string)).toEqual({ mode: "Auto" });
    });
  });

  it("disables mode changes in read-only mode", async () => {
    respond({ "/api/v1/info": { ...info, readOnly: true }, "/api/v1/workloads/shop/Deployment/api": workload() });
    renderPage();
    await screen.findByText("→ 126m");
    await waitFor(() => expect(screen.getByRole("radio", { name: /^auto/i })).toBeDisabled());
  });
});
