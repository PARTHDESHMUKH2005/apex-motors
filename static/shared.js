// Shared JS for all Apex Motors pages
function toggleMobileMenu() {
  const links = document.getElementById("navLinks");
  const hamburger = document.getElementById("hamburger");
  links.classList.toggle("mobile-open");
  hamburger.classList.toggle("active");
}

// Scroll reveal
document.addEventListener("DOMContentLoaded", () => {
  const revealElements = document.querySelectorAll(".reveal");
  const revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry, i) => {
        if (entry.isIntersecting) {
          setTimeout(() => entry.target.classList.add("visible"), i * 100);
          revealObserver.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.12 },
  );
  revealElements.forEach((el) => revealObserver.observe(el));
});
