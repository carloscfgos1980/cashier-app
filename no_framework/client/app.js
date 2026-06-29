const logPanel = document.getElementById("logPanel");
const billsList = document.getElementById("billsList");
const healthStatus = document.getElementById("healthStatus");
const changeResult = document.getElementById("changeResult");
let currentBills = [];

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  minimumFractionDigits: 0,
  maximumFractionDigits: 2,
});

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function formatCurrency(value) {
  return currencyFormatter.format(Number(value || 0));
}

function setLog(message, isError = false) {
  const text = typeof message === "string" ? message : JSON.stringify(message, null, 2);
  logPanel.innerHTML = `
    <div class="response-card ${isError ? "tone-error" : "tone-success"}">
      <div class="response-kicker">Server response</div>
      <div class="response-title">${isError ? "Something needs attention" : "Request completed"}</div>
      <p class="response-message">${escapeHtml(text)}</p>
    </div>
  `;
  logPanel.classList.toggle("error", isError);
}

function renderBillsSummary(bills) {
  if (!Array.isArray(bills) || bills.length === 0) {
    return `
      <div class="response-card tone-neutral">
        <div class="response-kicker">Bills synced</div>
        <div class="response-title">No bills stored yet</div>
        <p class="response-message">Add a bill to start building the change drawer.</p>
      </div>
    `;
  }

  const totalBills = bills.reduce((sum, bill) => sum + Number(bill.quantity || 0), 0);
  const totalCash = bills.reduce((sum, bill) => sum + Number(bill.quantity || 0) * Number(bill.denomination || 0), 0);

  return `
    <div class="response-card tone-success">
      <div class="response-kicker">Bills synced</div>
      <div class="response-title">${bills.length} denomination${bills.length === 1 ? "" : "s"} available</div>
      <p class="response-message">The register currently holds ${totalBills} bill${totalBills === 1 ? "" : "s"} worth ${formatCurrency(totalCash)}.</p>
      <div class="response-grid">
        ${bills
          .map(
            (bill) => `
              <div class="response-item">
                <span>${formatCurrency(bill.denomination)}</span>
                <strong>${bill.quantity} in stock</strong>
              </div>
            `,
          )
          .join("")}
      </div>
    </div>
  `;
}

function renderChangeSummary(changeBills, changeAmount) {
  const hasChange = Array.isArray(changeBills) && changeBills.length > 0;
  return `
    <div class="response-card ${hasChange ? "tone-success" : "tone-neutral"}">
      <div class="response-kicker">Change ready</div>
      <div class="response-title">${hasChange ? "Give back this breakdown" : "No change needed"}</div>
      <p class="response-message">${hasChange ? `Total change: ${formatCurrency(changeAmount)}.` : "The amount paid already covers the amount due."}</p>
      <div class="response-grid">
        ${hasChange
          ? changeBills
              .map(
                (bill) => `
                  <div class="response-item">
                    <span>${escapeHtml(bill.text || "")}</span>
                  </div>
                `,
              )
              .join("")
          : '<div class="response-item"><span>Ready</span><strong>No bills to dispense</strong></div>'}
      </div>
    </div>
  `;
}

async function requestJson(path, options = {}) {
  const response = await fetch(path, {
    cache: "no-store",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  const contentType = response.headers.get("content-type") || "";
  const payload = contentType.includes("application/json") ? await response.json() : await response.text();

  if (!response.ok) {
    throw new Error(payload?.error || response.statusText || "Request failed");
  }

  return payload;
}

function renderBills(bills) {
  if (!Array.isArray(bills) || bills.length === 0) {
    billsList.innerHTML = '<div class="list-item"><strong>No bills yet</strong><span class="meta">Add one to get started</span></div>';
    return;
  }

  billsList.innerHTML = bills
    .map((bill) => {
      return `
        <div class="list-item">
          <div>
            <strong>$${Number(bill.denomination).toFixed(2)}</strong>
          </div>
          <div><strong>${bill.quantity}</strong> units</div>
        </div>
      `;
    })
    .join("");
}

async function loadBills() {
  try {
    const bills = await requestJson(`/api/bills?t=${Date.now()}`);
    currentBills = Array.isArray(bills) ? bills : [];
    renderBills(bills);
    logPanel.innerHTML = renderBillsSummary(bills);
    logPanel.classList.remove("error");
  } catch (error) {
    setLog(error.message, true);
  }
}

async function checkHealth() {
  try {
    const status = await requestJson("/health");
    healthStatus.textContent = `${status.status} · ${status.message}`;
    healthStatus.classList.remove("error");
  } catch {
    healthStatus.textContent = "unavailable";
    healthStatus.classList.add("error");
  }
}

document.getElementById("billsForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const denomination = Number(form.get("denomination"));
  const inputQuantity = Number(form.get("quantity"));
  const existingBill = currentBills.find((bill) => Number(bill.denomination) === denomination);
  const payload = [
    {
      denomination,
      quantity: inputQuantity,
    },
  ];

  try {
    await requestJson("/api/bills", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    const existingBillIndex = currentBills.findIndex((bill) => Number(bill.denomination) === denomination);
    if (existingBillIndex >= 0) {
      currentBills[existingBillIndex].quantity = inputQuantity;
    } else {
      currentBills.push({ denomination, quantity: inputQuantity });
      currentBills.sort((a, b) => Number(b.denomination) - Number(a.denomination));
    }

    renderBills(currentBills);
    logPanel.innerHTML = renderBillsSummary(currentBills);
    logPanel.classList.remove("error");

    setLog(existingBill ? "Bill updated" : "Bill saved", false);
    event.currentTarget.reset();
    await loadBills();
  } catch (error) {
    setLog(error.message, true);
  }
});

document.getElementById("changeForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const payload = {
    amount_due: Number(form.get("amount_due")),
    amount_paid: Number(form.get("amount_paid")),
  };

  try {
    const result = await requestJson("/api/change", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    const changeAmount = Number(payload.amount_paid) - Number(payload.amount_due);
    changeResult.innerHTML = renderChangeSummary(result, changeAmount);
    setLog("Change calculated successfully", false);
  } catch (error) {
    changeResult.innerHTML = `<div class="response-card tone-error"><div class="response-kicker">Change request</div><div class="response-title">Unable to calculate</div><p class="response-message">${escapeHtml(error.message)}</p></div>`;
    setLog(error.message, true);
  }
});

checkHealth();
loadBills();