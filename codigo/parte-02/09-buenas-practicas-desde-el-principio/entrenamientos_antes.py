import math

d = [
    {"f": "2026-03-02", "km": 5.2, "min": 29, "ok": True},
    {"f": "2026-03-04", "km": 8.0, "min": 47, "ok": True},
    {"f": "2026-03-07", "km": 0, "min": 0, "ok": False},
    {"f": "2026-03-09", "km": 10.5, "min": 58, "ok": True},
]

def proc(l):
    t=0
    m=0
    mejor=None
    for x in l:
        if x["ok"] == True:
            t=t+x["km"]
            m=m+x["min"]
            r=x["min"]/x["km"]
            if mejor==None or r<mejor[1]:
                mejor=(x["f"],r)
    print("Total: "+str(round(t,1))+" km")
    print("Ritmo medio: "+str(round(m/t,2))+" min/km")
    print("Mejor ritmo: "+str(round(mejor[1],2))+" min/km el "+mejor[0])
    if t>=21.1:
        print("Esta semana has corrido más que una media maratón")

proc(d)
