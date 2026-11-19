"""Ejecuta los tests. Funciona también con `python -I`, que no añade la carpeta actual a sys.path."""

import sys
import unittest
from pathlib import Path

RAIZ = Path(__file__).resolve().parent
sys.path.insert(0, str(RAIZ))

if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    suite = unittest.defaultTestLoader.discover(str(RAIZ / "tests"), top_level_dir=str(RAIZ))
    resultado = unittest.TextTestRunner(verbosity=2).run(suite)
    sys.exit(0 if resultado.wasSuccessful() else 1)
