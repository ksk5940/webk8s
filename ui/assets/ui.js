const UI = {
  nsSel: () => document.getElementById("namespaceSelect"),
  resSel: () => document.getElementById("resourceSelect"),
  title: () => document.getElementById("title"),
  countInfo: () => document.getElementById("countInfo"),
  head: () => document.getElementById("tableHead"),
  body: () => document.getElementById("tableBody"),
  search: () => document.getElementById("searchInput"),
  drawer: () => document.getElementById("drawer"),
  drawerTitle: () => document.getElementById("drawerTitle"),
  drawerSubtitle: () => document.getElementById("drawerSubtitle"),
  tabContent: () => document.getElementById("tabContent"),
};

function openDrawer() {
  document.querySelector(".main-layout").style.gridTemplateColumns = "340px 1fr 540px";
  UI.drawer().classList.remove("closed");
}
function closeDrawer() {
  document.querySelector(".main-layout").style.gridTemplateColumns = "340px 1fr 0px";
  UI.drawer().classList.add("closed");
}
